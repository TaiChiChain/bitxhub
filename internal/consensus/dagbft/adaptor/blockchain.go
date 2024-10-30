package adaptor

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-kit/txpool"
	kittypes "github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	dagbft_common "github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/epochmgr"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	synccomm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/bcds/go-hpc-dagbft/app/demo/storage"
	"github.com/bcds/go-hpc-dagbft/common/config"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/events"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/bcds/go-hpc-dagbft/protocol/layer"
	"github.com/ethereum/go-ethereum/event"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go4.org/sort"
)

var _ layer.Ledger = (*BlockChain)(nil)

type LedgerConfig struct {
	useBls     bool
	ChainState *chainstate.ChainState

	GetBlockHeaderFunc func(height uint64) (*kittypes.BlockHeader, error)
}

type BlockChain struct {
	epochService  *epochmgr.EpochManager
	sync          synccomm.Sync
	crypto        layer.Crypto
	stores        map[types.Epoch]layer.StorageFactory
	txPool        txpool.TxPool[kittypes.Transaction, *kittypes.Transaction]
	dbBuilder     Builder
	batchVerifier *BatchVerifier

	ledgerConfig      *LedgerConfig
	narwhalConfig     config.Configs
	logger            logrus.FieldLogger
	closeCh           chan bool
	decisionCh        chan containers.Pair[*types.QuorumCheckpoint, chan<- *events.StableCommittedEvent]
	updatingCh        chan containers.Tuple[*types.QuorumCheckpoint, []*types.QuorumCheckpoint, chan<- *events.StateUpdatedEvent]
	executingCh       chan containers.Tuple[types.Height, *types.ConsensusOutput, chan<- *events.ExecutedEvent]
	waitEpochChangeCh chan struct{}

	sendReadyC          chan *Ready
	sendToExecuteCh     chan *consensustypes.CommitEvent
	notifyStateUpdateCh chan containers.Tuple[types.Height, *types.QuorumCheckpoint, chan<- *events.StateUpdatedEvent]
	metrics             *blockChainMetrics
	wg                  sync.WaitGroup
	batchWg             sync.WaitGroup
	batchLock           sync.Mutex
	getBlockFn          func(height uint64) (*kittypes.Block, error)

	AttestationFeed *event.Feed
}

func (b *BlockChain) LedgerType() config.LedgerType {
	return config.SingleLedger
}

func newBlockchain(precheck precheck.PreCheck, config *common.Config, narwhalConfig config.Configs, readyC chan *Ready, closeCh chan bool) (*BlockChain, error) {
	var (
		err   error
		store layer.Storage
	)
	storeBuilder := func(e types.Epoch, name string) layer.Storage {
		storePath := common.GenNodeDbPath(config, name, e)

		switch config.Repo.Config.Consensus.StorageType {
		case repo.ConsensusStorageTypeMinifile:
			store, err = common.OpenMinifile(storePath)
		case repo.ConsensusStorageTypeRosedb:
			store, err = common.OpenRosedb(storePath)
		case repo.ConsensusStorageTypePebble:
			store, err = common.OpenPebbleDb(storePath)
		}
		if err != nil {
			panic(err)
		}
		return store
	}
	epoch := config.ChainState.GetCurrentEpochInfo().Epoch
	stores := make(map[types.Epoch]layer.StorageFactory)
	stores[epoch] = NewFactory(epoch, storeBuilder)

	crp, err := NewCryptoImpl(config)
	if err != nil {
		return nil, err
	}
	return &BlockChain{
		dbBuilder:     storeBuilder,
		narwhalConfig: narwhalConfig,
		ledgerConfig:  &LedgerConfig{ChainState: config.ChainState, GetBlockHeaderFunc: config.GetBlockHeaderFunc},
		stores:        stores,
		crypto:        crp,
		epochService:  config.EpochStore,
		txPool:        config.TxPool,
		sync:          config.BlockSync,
		logger:        config.Logger,
		// init all channel
		closeCh:             closeCh,
		sendReadyC:          readyC,
		decisionCh:          make(chan containers.Pair[*types.QuorumCheckpoint, chan<- *events.StableCommittedEvent]),
		updatingCh:          make(chan containers.Tuple[*types.QuorumCheckpoint, []*types.QuorumCheckpoint, chan<- *events.StateUpdatedEvent]),
		executingCh:         make(chan containers.Tuple[types.Height, *types.ConsensusOutput, chan<- *events.ExecutedEvent], repo.ChannelSize),
		sendToExecuteCh:     make(chan *consensustypes.CommitEvent, 1),
		notifyStateUpdateCh: make(chan containers.Tuple[types.Height, *types.QuorumCheckpoint, chan<- *events.StateUpdatedEvent]),
		waitEpochChangeCh:   make(chan struct{}),
		metrics:             newBlockChainMetrics(),
		batchVerifier:       newBatchVerifier(precheck, config.Logger),
		getBlockFn:          config.GetBlockFunc,
		AttestationFeed:     new(event.Feed),
	}, nil
}

func (b *BlockChain) NotifyStateUpdate() <-chan containers.Tuple[types.Height, *types.QuorumCheckpoint, chan<- *events.StateUpdatedEvent] {
	return b.notifyStateUpdateCh
}

func (b *BlockChain) RecvStateUpdateEvent() <-chan *consensustypes.CommitEvent {
	return b.sendToExecuteCh
}

func (b *BlockChain) listenChainEvent() {
	for {
		select {
		case <-b.closeCh:
			return
		case execution := <-b.executingCh:
			stateHeight, output, evCh := execution.Unpack()
			metrics.Consensus2ExecutorBlockCounter.WithLabelValues(consensustypes.Dagbft).Inc()
			b.asyncExecute(output, stateHeight, evCh)
		case update := <-b.updatingCh:
			target, changes, evCh := update.Unpack()
			if len(changes) > 0 {
				b.logger.Infof("store epoch state at epoch %d", target.Epoch())
				b.wg.Add(1)
				go func() {
					defer b.wg.Done()
					lo.ForEach(changes, func(ckpt *types.QuorumCheckpoint, index int) {
						if ckpt != nil {
							epochChange := &consensustypes.DagbftQuorumCheckpoint{QuorumCheckpoint: ckpt}
							if err := b.epochService.StoreEpochState(epochChange); err != nil {
								b.logger.Errorf("failed to store epoch state at epoch %d: %s", ckpt.Epoch(), err)
							}
						}
					})
				}()
			}
			localHeight := b.ledgerConfig.ChainState.GetCurrentCheckpointState().Height
			targetHeight := target.Checkpoint().ExecuteState().StateHeight()
			latestBlockHash := b.ledgerConfig.ChainState.GetCurrentCheckpointState().Digest

			if localHeight >= targetHeight {
				localHeader, err := b.ledgerConfig.GetBlockHeaderFunc(targetHeight)
				if err != nil {
					panic(err)
				}
				if localHeader.Hash().String() != target.StateRoot().String() {
					panic(fmt.Errorf("local state root[%s] is not equal to checkpoint state root[%s]",
						localHeader.Hash().String(), target.StateRoot().String()))
				} else {
					b.wg.Wait()
					b.logger.Infof("node state have reached target height: %d, ignore updating...", targetHeight)
					stateUpdated := &events.StateUpdatedEvent{
						Checkpoint: target,
						Updated:    false,
					}
					channel.SafeSend(evCh, stateUpdated, b.closeCh)
					continue
				}
			}

			b.logger.Infof("node start state updating to target height: %d", targetHeight)
			stateResult := containers.Pack3(target.Height(), target, evCh)
			channel.SafeSend(b.notifyStateUpdateCh, stateResult, b.closeCh)
			if err := b.update(localHeight, latestBlockHash, target, changes); err != nil {
				panic(fmt.Errorf("failed to update: %w", err))
			}

		case decision := <-b.decisionCh:
			checkpoint, evCh := decision.Unpack()
			b.checkpoint(checkpoint)

			stableCommitted := &events.StableCommittedEvent{
				Checkpoint: checkpoint,
			}
			channel.SafeSend(evCh, stableCommitted, b.closeCh)
		}
	}
}

func (b *BlockChain) GetLedgerState(height *types.Height) (*types.ExecuteState, error) {
	var blockHeight uint64
	if height == nil {
		blockHeight = b.ledgerConfig.ChainState.ChainMeta.Height
	} else {
		blockHeight = *height
	}

	header, err := b.ledgerConfig.GetBlockHeaderFunc(blockHeight)
	if err != nil {
		return nil, err
	}

	return &types.ExecuteState{
		ExecuteState: protos.ExecuteState{
			Height:    blockHeight,
			StateRoot: header.Hash().String(),
		},
		Reconfigured: common.NeedChangeEpoch(blockHeight, b.ledgerConfig.ChainState.GetCurrentEpochInfo()),
	}, nil
}

// todo: add workers in Chain state, support dynamic adjustment of worker parameters.
func (b *BlockChain) GetLedgerValidators() (protocol.Validators, error) {
	var validators []*protos.Validator
	validators = lo.Map(b.ledgerConfig.ChainState.ValidatorSet, func(item chainstate.ValidatorInfo, index int) *protos.Validator {
		nodeInfo, err := b.ledgerConfig.ChainState.GetNodeInfo(item.ID)
		if err != nil {
			b.logger.Warningf("failed to get node info: %v", err)
			return nil
		}
		var pubBytes []byte
		if b.ledgerConfig.useBls {
			pubBytes, err = nodeInfo.ConsensusPubKey.Marshal()
		} else {
			pubBytes, err = nodeInfo.P2PPubKey.Marshal()
		}
		if err != nil {
			b.logger.Warningf("failed to marshal pub key: %v", err)
			return nil
		}
		return &protos.Validator{
			Hostname:    nodeInfo.Primary,
			PubKey:      pubBytes,
			ValidatorId: uint32(item.ID),
			VotePower:   uint64(item.ConsensusVotingPower),
			Workers:     nodeInfo.Workers,
		}
	})
	return validators, nil
}

// todo(lrx): modify this function when the ledger version is upgraded
func (b *BlockChain) GetLedgerAlgoVersion() (string, error) {
	return "DagBFT@1.0", nil
}

func (b *BlockChain) recordBatchMetrics(output *types.ConsensusOutput, height types.Height) {
	now := time.Now().UnixNano()
	var maxLatency, minLatency, totalLatency types.TimestampNs = 0, now, 0
	batchCount := 0
	for _, batches := range output.Batches {
		for _, batch := range batches {
			// compute latency for metrics
			latency := now - batch.Timestamp()
			if latency > maxLatency {
				maxLatency = latency
			}
			if latency < minLatency {
				minLatency = latency
			}
			totalLatency += latency
			batchCount++
			b.logger.WithFields(logrus.Fields{
				"maxLatency": time.Duration(maxLatency),
				"minLatency": time.Duration(minLatency),
				"avgLatency": time.Duration(float64(totalLatency) / float64(batchCount)),
			}).Infof("[DagBFT.Ledger] Execute Batch %s with %d txs in height: %d",
				batch.Digest(), len(batch.GetTransactions()), height)
		}
	}

	if batchCount > 0 {
		metrics.BatchCommitLatency.With(prometheus.Labels{"consensus": consensustypes.Dagbft, "type": "max"}).Observe(time.Duration(maxLatency).Seconds())
		metrics.BatchCommitLatency.With(prometheus.Labels{"consensus": consensustypes.Dagbft, "type": "min"}).Observe(time.Duration(minLatency).Seconds())
		metrics.BatchCommitLatency.With(prometheus.Labels{"consensus": consensustypes.Dagbft, "type": "avg"}).Observe(time.Duration(totalLatency).Seconds() / float64(batchCount))
	}
}

func (b *BlockChain) readLedgerEpoch() uint64 {
	return b.ledgerConfig.ChainState.GetCurrentEpochInfo().Epoch
}

func (b *BlockChain) Execute(output *types.ConsensusOutput, height types.Height, eventCh chan<- *events.ExecutedEvent) error {
	b.logger.Infof("Execute Output %d at height %d, batches %d, txs: %d",
		output.CommitInfo.CommitSequence, height, output.BatchCount(), output.TransactionCount())
	b.recordBatchMetrics(output, height)
	transactionCount := output.TransactionCount()
	outputEpoch := output.Epoch()
	ledgerEpoch := b.readLedgerEpoch()
	if ledgerEpoch != outputEpoch {
		b.logger.Warningf("[DagBFT.Ledger] Discard mismatched epoch outputs, ledger: %d, output: %d", ledgerEpoch, outputEpoch)
		b.metrics.discardedTransactions.WithLabelValues("epoch_mismatched").Add(float64(transactionCount))
		// TODO: send sp event or return error but not nil, and drop following executions in the core
		return nil
	}
	isConfigured := common.NeedChangeEpoch(height, b.ledgerConfig.ChainState.GetCurrentEpochInfo())
	channel.SafeSend(b.executingCh, containers.Pack3(height, output, eventCh), b.closeCh)
	if isConfigured {
		start := time.Now()
		select {
		case <-b.waitEpochChangeCh:
			metrics.WaitEpochTime.WithLabelValues(consensustypes.Dagbft).Observe(time.Since(start).Seconds())
			b.logger.WithFields(logrus.Fields{
				"duration": time.Since(start),
			}).Info("wait epoch change done")
			return nil
		case <-b.closeCh:
			return nil
		//todo: configurable timeout
		case <-time.After(60 * time.Second):
			return fmt.Errorf("wait epoch change timeout")
		}
	}

	return nil
}

func (b *BlockChain) asyncExecute(output *types.ConsensusOutput, height types.Height, executedCh chan<- *events.ExecutedEvent) {
	recvConsensusOutPut := time.Now().UnixNano()
	validTxs := b.filterValidTxs(output.Batches)
	b.logger.Debugf("valid txs in execute: %d", len(validTxs))
	ready := &Ready{
		Txs:                    validTxs,
		Height:                 height,
		RecvConsensusTimestamp: recvConsensusOutPut,
		Timestamp:              output.Timestamp(),
		ProposerNodeID:         uint64(output.CommitInfo.Leader.GetHeader().GetAuthorId()),
		ExecutedCh:             executedCh,
		CommitState: &types.CommitState{
			CommitState: protos.CommitState{
				Sequence:     output.CommitSequence(),
				CommitDigest: output.CommitInfo.Digest().String(),
			},
		},
	}
	if common.NeedChangeEpoch(height, b.ledgerConfig.ChainState.GetCurrentEpochInfo()) {
		ready.EpochChangedCh = b.waitEpochChangeCh
	}
	channel.SafeSend(b.sendReadyC, ready, b.closeCh)
}

func (b *BlockChain) filterValidTxs(batches [][]*types.Batch) []*kittypes.Transaction {
	flattenBatches := lo.Flatten(batches)

	// 1. sort batch by batchTime, ensure that the batches for the same worker are sorted by time
	sort.Slice(flattenBatches, func(i, j int) bool {
		return flattenBatches[i].Batch.MetaData.Timestamp < flattenBatches[j].Batch.MetaData.Timestamp
	})

	validTxs := make([][]*kittypes.Transaction, len(flattenBatches))
	start := time.Now()
	txCount := atomic.Uint64{}
	// 2.1 precheck txs async (including basic check、 verify Signature、 check tx data validity)
	b.batchWg.Add(len(flattenBatches))
	// todo: add workerID filed in Batch meta, and use it to ignore prechcked txs
	lo.ForEach(flattenBatches, func(val *types.Batch, index int) {
		go func(batch *types.Batch, i int) {
			defer b.batchWg.Done()
			txCount.Add(uint64(len(batch.GetTransactions())))
			b.logger.Infof("start verify batch: %s, txs: %d", batch.Digest().String(), len(batch.Transactions))
			txs := b.batchVerifier.verifyBatchTransactions(batch.Transactions, b.metrics)
			b.batchLock.Lock()
			validTxs[i] = txs
			b.batchLock.Unlock()
		}(val, index)
	})
	b.batchWg.Wait()

	// 2.2 flatten valid txs from each batch
	ouput := lo.Flatten(validTxs)

	// 3. filter duplicated txs
	uniqTxs := lo.UniqBy(ouput, func(tx *kittypes.Transaction) string {
		return tx.GetHash().String()
	})

	b.metrics.batchVerifyLatency.Observe(time.Since(start).Seconds())
	b.metrics.batchCounter.Add(float64(len(batches)))
	b.metrics.batchTxsCounter.Add(float64(txCount.Load()))
	return uniqTxs
}

func (b *BlockChain) Checkpoint(checkpoint *types.QuorumCheckpoint, eventCh chan<- *events.StableCommittedEvent) error {
	b.logger.Infof("Checkpoint at %d,%d", checkpoint.Checkpoint().CommitState().CommitSequence(), checkpoint.Checkpoint().ExecuteState().StateHeight())
	channel.SafeSend(b.decisionCh, containers.Pack2(checkpoint, eventCh), b.closeCh)
	return nil
}

func (b *BlockChain) checkpoint(checkpoint *types.QuorumCheckpoint) {
	if checkpoint.Height() <= b.ledgerConfig.ChainState.GetCurrentCheckpointState().Height {
		b.logger.Warningf("ignore checkpoint at height %d", checkpoint.Height())
		return
	}
	if checkpoint.Checkpoint().EndsEpoch() {
		//todo: update validators and crypto validatorVerifier
		epochChange := &consensustypes.DagbftQuorumCheckpoint{QuorumCheckpoint: checkpoint}
		if err := b.epochService.StoreEpochState(epochChange); err != nil {
			b.logger.Errorf("failed to store epoch state at epoch %d: %v", checkpoint.Epoch(), err)
		}
	}
	proof := &consensustypes.DagbftQuorumCheckpoint{QuorumCheckpoint: checkpoint}
	proofData, err := proof.Marshal()
	if err != nil {
		b.logger.Errorf("failed to marshal proof data: %v", err)
		return
	}
	validators, err := b.GetLedgerValidators()
	if err != nil {
		b.logger.Errorf("failed to get validators: %v", err)
		return
	}

	b.ledgerConfig.ChainState.UpdateCheckpoint(&pb.QuorumCheckpoint{
		Epoch: checkpoint.Epoch(),
		State: &pb.ExecuteState{
			Height: checkpoint.Checkpoint().ExecuteState().StateHeight(),
			Digest: checkpoint.Checkpoint().ExecuteState().StateRoot().String(),
		},
		Proof: proofData,
		ValidatorSet: lo.Map(validators, func(v *protos.Validator, _ int) *pb.QuorumCheckpoint_Validator {
			p2pId, err := b.ledgerConfig.ChainState.GetP2PIDByNodeID(uint64(v.ValidatorId))
			if err != nil {
				b.logger.Errorf("failed to get p2p id: %v", err)
				return nil
			}
			return &pb.QuorumCheckpoint_Validator{
				Id:      uint64(v.ValidatorId),
				P2PId:   p2pId,
				Primary: v.Hostname,
				Workers: v.Workers,
			}
		}),
	})

	if err = dagbft_common.PostAttestationEvent(checkpoint, b.AttestationFeed, b.getBlockFn); err != nil {
		b.logger.Errorf("failed to post attestation event: %v", err)
	}
}

func (b *BlockChain) StateUpdate(checkpoint *types.QuorumCheckpoint, eventCh chan<- *events.StateUpdatedEvent, epochChanges ...*types.QuorumCheckpoint) error {
	b.logger.Infof("Update State to sequence: %d, height: %d", checkpoint.Checkpoint().CommitState().CommitSequence(), checkpoint.Checkpoint().ExecuteState().StateHeight())

	channel.SafeSend(b.updatingCh, containers.Pack3(checkpoint, epochChanges, eventCh), b.closeCh)
	return nil
}

func (b *BlockChain) update(localHeight types.Height, latestBlockHash string, checkpoint *types.QuorumCheckpoint, epochChanges []*types.QuorumCheckpoint) error {
	b.metrics.syncChainCounter.Inc()
	targetHeight := checkpoint.Checkpoint().ExecuteState().StateHeight()

	peerM := dagbft_common.GetRemotePeers(epochChanges, b.ledgerConfig.ChainState)

	syncTaskDoneCh := make(chan error, 1)
	if err := retry.Retry(func(attempt uint) error {
		params := &synccomm.SyncParams{
			Peers:           peerM,
			LatestBlockHash: latestBlockHash,
			// ensure sync remote count including at least one correct node
			Quorum:       common.CalFaulty(uint64(len(peerM) + 1)),
			CurHeight:    localHeight + 1,
			TargetHeight: targetHeight,
			QuorumCheckpoint: &consensustypes.DagbftQuorumCheckpoint{
				QuorumCheckpoint: checkpoint,
			},
			EpochChanges: lo.Map(epochChanges, func(qckpt *types.QuorumCheckpoint, index int) *consensustypes.EpochChange {
				return &consensustypes.EpochChange{
					QuorumCheckpoint: &consensustypes.DagbftQuorumCheckpoint{QuorumCheckpoint: checkpoint},
				}
			}),
		}
		b.logger.WithFields(logrus.Fields{
			"target":       params.TargetHeight,
			"target_hash":  params.QuorumCheckpoint.GetStateDigest(),
			"start":        params.CurHeight,
			"epochChanges": epochChanges,
		}).Info("State update start")
		err := b.sync.StartSync(params, syncTaskDoneCh)
		if err != nil {
			b.logger.Infof("start sync failed[local:%b, target:%b]: %s", localHeight, targetHeight, err)
			return err
		}
		return nil
	}, strategy.Limit(5), strategy.Wait(500*time.Microsecond)); err != nil {
		panic(fmt.Errorf("retry start sync failed: %v", err))
	}

	b.wg.Wait()
	var stateUpdatedCheckpoint *consensustypes.Checkpoint
	// wait for the sync to finish
	for {
		select {
		case <-b.closeCh:
			b.logger.Info("state update is canceled!!!!!!")
			return nil
		case syncErr := <-syncTaskDoneCh:
			if syncErr != nil {
				return syncErr
			}
		case data := <-b.sync.Commit():
			endSync := false
			blockCache, ok := data.([]synccomm.CommitData)
			if !ok {
				panic("state update failed: invalid commit data")
			}

			b.logger.Infof("fetch chunk: start: %d, end: %d", blockCache[0].GetHeight(), blockCache[len(blockCache)-1].GetHeight())

			for _, commitData := range blockCache {
				// if the block is the target block, we should resign the stateUpdatedCheckpoint in CommitEvent
				// and send the quitSync signal to sync module
				if commitData.GetHeight() == targetHeight {
					stateUpdatedCheckpoint = &consensustypes.Checkpoint{
						Epoch:  checkpoint.Epoch(),
						Height: checkpoint.Height(),
						Digest: checkpoint.Checkpoint().ExecuteState().StateRoot().String(),
					}
					endSync = true
				}
				block, ok := commitData.(*synccomm.BlockData)
				if !ok {
					panic("state update failed: invalid commit data")
				}
				commitEvent := &consensustypes.CommitEvent{
					Block:                  block.Block,
					StateUpdatedCheckpoint: stateUpdatedCheckpoint,
				}
				channel.SafeSend(b.sendToExecuteCh, commitEvent, b.closeCh)

				if endSync {
					b.logger.Infof("State update finished, target height: %d", targetHeight)
					return nil
				}
			}
		}
	}
}

func (b *BlockChain) GetEpoch() types.Epoch {
	return b.ledgerConfig.ChainState.GetCurrentEpochInfo().Epoch
}

func (b *BlockChain) GetEpochCrypto() layer.Crypto {
	return b.crypto
}

func (b *BlockChain) GetEpochConfig(epoch types.Epoch) config.Configs {
	return b.narwhalConfig
}

func (b *BlockChain) GetEpochStorage(epoch types.Epoch) layer.StorageFactory {
	b.logger.Infof("[BlockChain] Get Storage of epoch %d", epoch)
	sf, ok := b.stores[epoch]
	if !ok {
		sf = storage.NewFactory(epoch, b.dbBuilder)
		b.stores[epoch] = sf
	}
	return sf
}

// GetEpochCheckpoint query the configured checkpoint by given epoch in ledger.
// if EpochPeriod is 100, we will query the checkpoint in the following way:
//
//	epoch:1 -> genesis checkpoint(height:0)
//	epoch:2 -> configured checkpoint(height:99)
//	epoch:3 -> configured checkpoint(height:199)
//	...
func (b *BlockChain) GetEpochCheckpoint(epoch *types.Epoch) *types.QuorumCheckpoint {
	if epoch == nil {
		latestEpoch := b.ledgerConfig.ChainState.GetCurrentEpochInfo().Epoch
		epoch = &latestEpoch
	}
	raw, err := b.epochService.ReadEpochState(*epoch)
	if err != nil {
		b.logger.Errorf("failed to read epoch %d quorum chkpt: %v", epoch, err)
		return nil
	}
	data, ok := raw.(*consensustypes.DagbftQuorumCheckpoint)
	if !ok {
		b.logger.Errorf("failed to read epoch %d quorum chkpt: type assertion failed: %v", epoch, reflect.TypeOf(raw))
		return nil
	}
	return data.QuorumCheckpoint
}
