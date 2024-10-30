package dagbft

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	"github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/data_syncer"
	"github.com/axiomesh/axiom-ledger/internal/consensus/txcache"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/adaptor"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	dagbft "github.com/bcds/go-hpc-dagbft"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	dagevents "github.com/bcds/go-hpc-dagbft/common/types/events"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/ethereum/go-ethereum/event"
	"github.com/sirupsen/logrus"
)

func init() {
	repo.Register(repo.ConsensusTypeDagBft, true)
}

type executeEvent struct {
	containers.Tuple[*dagtypes.CommitState, chan<- *dagevents.ExecutedEvent, chan<- struct{}]
}

type Node struct {
	nodeConfig              *Config
	networkFactory          *adaptor.NetworkFactory
	stack                   *adaptor.DagBFTAdaptor
	currentEpochInfo        types.EpochInfo
	txpool                  txpool.TxPool[types.Transaction, *types.Transaction]
	engine                  dagbft.DagBFT
	dataSyncer              *data_syncer.Node
	txPreCheck              precheck.PreCheck
	recvTxCh                chan *consensustypes.TxWithResp
	recvRemoteTxsCh         chan [][]byte
	recvNotifyStateUpdateCh <-chan containers.Tuple[dagtypes.Height, *dagtypes.QuorumCheckpoint, chan<- *dagevents.StateUpdatedEvent]
	recvCommitEventCh       <-chan *consensustypes.CommitEvent

	txCache           txcache.TxCache
	reportBatchResult bool

	primary dagtypes.Host

	worker      dagtypes.Host // todo: support multiple workers
	readyC      chan *adaptor.Ready
	blockC      chan *consensustypes.CommitEvent
	executedMap sync.Map

	inStateUpdate atomic.Bool
	updatedResult containers.Tuple[dagtypes.Height, *dagtypes.QuorumCheckpoint, chan<- *dagevents.StateUpdatedEvent]

	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	closeCh chan bool
	crashCh chan bool

	logger        logrus.FieldLogger
	txFeed        event.Feed
	mockBlockFeed event.Feed
}

func NewNode(config *common.Config) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	crashCh := make(chan bool)
	closeCh := make(chan bool)
	dagConfig, err := GenerateDagBftConfig(config)

	readyC := make(chan *adaptor.Ready)
	check := precheck.NewTxPreCheckMgr(ctx, config)
	dagAdaptor, err := adaptor.NewAdaptor(check, config, dagConfig.NetworkConfig, readyC, dagConfig.DAGConfigs, ctx, closeCh)
	if err != nil {
		cancel()
		return nil, err
	}

	recvRemoteTxsCh := make(chan [][]byte, common.MaxChainSize)
	networkFactory := adaptor.NewNetworkFactory(config, ctx, recvRemoteTxsCh)

	recvNotifyStateUpdateCh := make(chan containers.Tuple[dagtypes.Height, *dagtypes.QuorumCheckpoint, chan<- *dagevents.StateUpdatedEvent])
	recvCommitEventCh := make(chan *consensustypes.CommitEvent)

	var (
		dataSyncer      *data_syncer.Node
		consensusEngine dagbft.DagBFT
	)

	if config.ChainState.IsDataSyncer {
		dataSyncer, err = data_syncer.NewNode(config, dagAdaptor.GetCryptoVerifier(), config.Logger, readyC, recvCommitEventCh, recvNotifyStateUpdateCh)
		if err != nil {
			cancel()
			return nil, err
		}
	} else {
		consensusEngine = dagbft.New(networkFactory, dagAdaptor.GetLedger(), dagConfig.Logger, dagConfig.MetricsProv, crashCh)
	}

	if err != nil {
		cancel()
		return nil, err
	}

	return &Node{
		blockC:         make(chan *consensustypes.CommitEvent),
		nodeConfig:     dagConfig,
		networkFactory: networkFactory,
		txpool:         config.TxPool,
		txPreCheck:     check,
		engine:         consensusEngine,
		dataSyncer:     dataSyncer,
		readyC:         readyC,
		logger:         config.Logger,
		stack:          dagAdaptor,
		ctx:            ctx,
		cancel:         cancel,
		txCache:        txcache.NewTxCache(config.Repo.ConsensusConfig.Dagbft.BatchTimeout.ToDuration(), uint64(config.Repo.ConsensusConfig.Dagbft.MaxBatchCount), config.Logger),

		recvTxCh:                make(chan *consensustypes.TxWithResp, common.MaxChainSize),
		closeCh:                 closeCh,
		crashCh:                 crashCh,
		primary:                 config.ChainState.SelfNodeInfo.Primary,
		worker:                  config.ChainState.SelfNodeInfo.Workers[0],
		executedMap:             sync.Map{},
		inStateUpdate:           atomic.Bool{},
		currentEpochInfo:        config.ChainState.GetCurrentEpochInfo(),
		reportBatchResult:       config.Repo.ConsensusConfig.Dagbft.ReportBatchResult,
		recvCommitEventCh:       recvCommitEventCh,
		recvNotifyStateUpdateCh: recvNotifyStateUpdateCh,
		recvRemoteTxsCh:         recvRemoteTxsCh,
	}, nil
}

func (n *Node) Start() error {
	n.txpool.Init(txpool.ConsensusConfig{
		SelfID: n.nodeConfig.ChainState.SelfNodeInfo.ID,
	})

	err := n.networkFactory.Start(n.engine)
	if err != nil {
		return err
	}

	n.stack.Start()

	if n.dataSyncer != nil {
		err = n.dataSyncer.Start()
		if err != nil {
			return err
		}
	} else {
		err = n.engine.Start()
		if err != nil {
			return err
		}
	}

	n.txPreCheck.Start()

	err = n.txpool.Start()
	if err != nil {
		return err
	}

	n.wg.Add(2)
	go n.listenLocalEvent()
	go n.listenConsensusEngineEvent()
	n.txCache.Start()

	n.logger.Info("dagbft consensus started")
	return nil
}

func (n *Node) listenLocalEvent() {
	defer n.wg.Done()
	for {
		select {
		case <-n.closeCh:
			return

		case crash := <-n.crashCh:
			if crash {
				n.logger.Errorf("consensus crashed,, see details in node log")
				n.Stop()
			}

		case wrapTx := <-n.recvTxCh:
			ev := &consensustypes.UncheckedTxEvent{
				EventType: consensustypes.LocalTxEvent,
				Event:     wrapTx,
			}
			n.txPreCheck.PostUncheckedTxEvent(ev)
		case txs := <-n.txPreCheck.CommitValidTxs():
			if !txs.Local {
				panic("only support local tx in dagbft mode!!!")
			}
			precheck.RespLocalTx(txs.LocalCheckRespCh, nil)
			err := n.txpool.AddLocalTx(txs.Txs[0])
			precheck.RespLocalTx(txs.LocalPoolRespCh, err)

		case rawTxs := <-n.txpool.ReceivePriorityTxs():
			n.logger.Infof("recv %d txs from txpool", len(rawTxs))
			lo.ForEach(rawTxs, func(tx *types.Transaction, _ int) {
				n.txCache.PostTx(tx)
			})

		case txSet := <-n.txCache.CommitTxSet():
			marshallTxs := lo.Map(txSet, func(tx *types.Transaction, _ int) protocol.Transaction {
				data, err := tx.Marshal()
				if err != nil {
					panic(err)
				}
				return data
			})

			// todo(lrx): handle error response, notify txpool revert nonce
			n.logger.WithFields(logrus.Fields{
				"count": len(marshallTxs),
			}).Info("send tx to consensus engine")
			respCh := make(chan any, 1)
			n.receiveTransactions(marshallTxs, respCh)
			metrics.SendTx2ConsensusCounter.Add(float64(len(txSet)))
			go func() {
				select {
				case <-n.closeCh:
					return
				case res := <-respCh:
					if err, ok := res.(error); ok {
						n.logger.WithError(err).Error("failed to send tx")
					} else {
						d := res.(dagtypes.BatchDigest)
						n.logger.Info("succeed to send txs, batchDigest:", d)
					}
				}
			}()

		}
	}
}

func (n *Node) receiveTransactions(transactions []protocol.Transaction, resultCh chan any) {
	const BatchTimeout = time.Minute

	if n.dataSyncer != nil {
		n.dataSyncer.SendNewTxSet(transactions)
		return
	}

	respCh := n.engine.GenerateBatch(n.worker, transactions)

	if !n.reportBatchResult {
		return
	}

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		timer := time.NewTimer(BatchTimeout)
		defer timer.Stop()

		select {
		case <-n.closeCh:

		case <-timer.C:
			err := fmt.Errorf("batch timeout for %v", BatchTimeout)
			channel.SafeSend[any, bool](resultCh, err, n.closeCh)

		case resp := <-respCh:
			batchMeta, err := resp.Unpack()
			var result any
			if err != nil {
				result = err
			} else {
				result = batchMeta.Digest
			}
			channel.SafeSend[any, bool](resultCh, result, n.closeCh)
		}
	}()

}

func (n *Node) Stop() {
	if n.dataSyncer != nil {
		n.dataSyncer.Stop()
	} else {
		n.engine.Stop()
	}
	close(n.closeCh)
	n.cancel()
	n.wg.Wait()
	n.logger.Info("dagbft consensus stopped")
}

func (n *Node) Prepare(tx *types.Transaction) error {
	defer n.txFeed.Send([]*types.Transaction{tx})
	txWithResp := &consensustypes.TxWithResp{
		Tx:      tx,
		CheckCh: make(chan *consensustypes.TxResp, 1),
		PoolCh:  make(chan *consensustypes.TxResp, 1),
	}

	channel.SafeSend(n.recvTxCh, txWithResp, n.closeCh)
	precheckResp, ok := channel.SafeRecv(txWithResp.CheckCh, n.closeCh)
	if !ok {
		n.logger.Warning("[DagBft] Exit Prepare tx for shutdown")
		return nil
	}
	if !precheckResp.Status {
		return errors.Wrap(consensustypes.ErrorPreCheck, precheckResp.ErrorMsg)
	}

	poolResp, ok := channel.SafeRecv(txWithResp.PoolCh, n.closeCh)
	if !ok {
		n.logger.Warning("[DagBft] Exit Prepare tx for shutdown")
		return nil
	}
	if !poolResp.Status {
		return errors.Wrap(consensustypes.ErrorAddTxPool, poolResp.ErrorMsg)
	}

	return nil
}

func (n *Node) Commit() chan *consensustypes.CommitEvent {
	return n.blockC
}

func (n *Node) Step(_ []byte) error {
	return nil
}

func (n *Node) Ready() error {
	if n.dataSyncer != nil {
		if !n.dataSyncer.IsNormal() {
			return fmt.Errorf("data syncer is abnormal")
		}
		return nil
	}
	status := n.engine.ReadStatus()
	if status.Primary.Status.String() == "Normal" {
		return nil
	}
	return fmt.Errorf("%s", status.Primary.Status.String())
}

func (n *Node) ReportState(height uint64, blockHash *types.Hash, txHashList []*events.TxPointer, stateUpdatedCheckpoint *consensustypes.Checkpoint, needRemoveTxs bool, commitSequence uint64) {
	reconfigured := common.NeedChangeEpoch(height, n.currentEpochInfo)

	// if state updated, notify state updated to consensus
	if n.inStateUpdate.Load() {
		if stateUpdatedHeight, quorumCkpt, respCh := n.updatedResult.Unpack(); stateUpdatedHeight == height {
			channel.SafeSend(respCh, &dagevents.StateUpdatedEvent{Checkpoint: quorumCkpt, Updated: true}, n.closeCh)
			n.logger.WithFields(logrus.Fields{
				"height": height,
				"hash":   blockHash,
			}).Infof("state updated")
			n.inStateUpdate.Store(false)
			n.updatedResult = containers.Tuple[dagtypes.Height, *dagtypes.QuorumCheckpoint, chan<- *dagevents.StateUpdatedEvent]{}
		}
		return
	}

	if executed, ok := n.executedMap.Load(commitSequence); ok {
		commitState, respCh, waitCh := executed.(*executeEvent).Unpack()
		committedEvent := &dagevents.ExecutedEvent{
			CommitState: commitState,
			ExecutedState: &dagtypes.ExecuteState{
				ExecuteState: protos.ExecuteState{
					Height:    height,
					StateRoot: blockHash.String(),
				},
				Reconfigured: reconfigured,
			},
		}
		metrics.ExecutedBlockCounter.WithLabelValues(consensustypes.Dagbft).Inc()
		channel.SafeSend(respCh, committedEvent, n.closeCh)
		// if reconfigured, notify epoch changed to consensus
		if reconfigured {
			channel.SafeSend(waitCh, struct{}{}, n.closeCh)
			n.updateEpochInfo()
		}
		n.executedMap.Delete(commitSequence)
		n.logger.WithFields(logrus.Fields{
			"height":       height,
			"hash":         blockHash,
			"sequence":     commitSequence,
			"reConfigured": reconfigured,
		}).Infof("report state")
	} else {
		n.logger.Errorf("no executed event for sequence %d", commitSequence)
	}
}

func (n *Node) Quorum(N uint64) uint64 {
	return common.CalQuorum(N)
}

func (n *Node) GetLowWatermark() uint64 {
	checkpoint := n.nodeConfig.ChainState.GetCurrentCheckpointState()
	return checkpoint.GetHeight()
}

func (n *Node) SubscribeTxEvent(events chan<- []*types.Transaction) event.Subscription {
	return n.txFeed.Subscribe(events)
}

func (n *Node) SubscribeMockBlockEvent(ch chan<- events.ExecutedEvent) event.Subscription {
	return n.mockBlockFeed.Subscribe(ch)
}

func (n *Node) SubscribeAttestationEvent(ch chan<- events.AttestationEvent) event.Subscription {
	if n.dataSyncer != nil {
		return n.dataSyncer.AttestationFeed.Subscribe(ch)
	}
	return n.stack.Chain.AttestationFeed.Subscribe(ch)
}

func (n *Node) listenConsensusEngineEvent() {
	defer n.wg.Done()
	for {
		select {
		case <-n.closeCh:
			return
		case res := <-n.recvNotifyStateUpdateCh:
			n.updatedResult = res
			n.inStateUpdate.Store(true)
		case ev := <-n.recvCommitEventCh:
			if n.inStateUpdate.Load() {
				n.logger.Infof("send [block %d] to executor", ev.Block.Height())
				channel.SafeSend(n.blockC, ev, n.closeCh)
			}
		case r := <-n.readyC:
			block := &types.Block{
				Header: &types.BlockHeader{
					Number:         r.Height,
					Timestamp:      r.Timestamp / int64(time.Second),
					ProposerNodeID: r.ProposerNodeID,
				},
				Transactions: r.Txs,
			}
			commitEvent := &consensustypes.CommitEvent{
				Block:             block,
				RecvConsensusTime: r.RecvConsensusTimestamp,
				CommitSequence:    r.CommitState.CommitSequence(),
			}

			n.logger.Infof("send [block %d commitState %d] to executor", r.Height, r.CommitState.CommitSequence())
			n.executedMap.Store(r.CommitState.CommitSequence(), &executeEvent{containers.Pack3(r.CommitState, r.ExecutedCh, r.EpochChangedCh)})
			metrics.Consensus2ExecuteBlockTime.WithLabelValues(consensustypes.Dagbft).Observe(float64(time.Now().UnixNano()-r.RecvConsensusTimestamp) / float64(time.Second))
			channel.SafeSend(n.blockC, commitEvent, n.closeCh)
		}
	}
}

func (n *Node) listenTxsBroadcastMsg() {
	defer n.wg.Done()
	for {
		select {
		case <-n.closeCh:
			return
		case txs := <-n.recvRemoteTxsCh:
			n.submitTxsFromRemote(txs)
		}
	}
}

func (n *Node) submitTxsFromRemote(txs [][]byte) {
	var requests []*types.Transaction
	for _, item := range txs {
		tx := &types.Transaction{}
		if err := tx.RbftUnmarshal(item); err != nil {
			n.logger.Error(err)
			continue
		}
		requests = append(requests, tx)
	}

	n.txFeed.Send(requests)
	ev := &consensustypes.UncheckedTxEvent{
		EventType: consensustypes.RemoteTxEvent,
		Event:     requests,
	}
	n.txPreCheck.PostUncheckedTxEvent(ev)
}

func (n *Node) updateEpochInfo() {
	n.currentEpochInfo = n.nodeConfig.ChainState.GetCurrentEpochInfo()
}
