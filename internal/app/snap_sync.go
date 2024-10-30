package app

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	rbft "github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	consensuscommon "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/epochmgr"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type snapMeta struct {
	snapBlockHeader  *types.BlockHeader
	startHeight      uint64
	latestBlockHash  string
	snapPersistEpoch uint64
	snapPeers        []*common.Node
}

type snapSyncManager struct {
	snapMeta         *snapMeta
	syncMgr          common.Sync
	chainLedger      ledger.ChainLedger
	stateVl          ledger.StateLedger
	rep              *repo.Repo
	epochMgr         *epochmgr.EpochManager
	logger           logrus.FieldLogger
	wg               sync.WaitGroup
	closeCh          chan struct{}
	ctx              context.Context
	epochChangeCache []*consensustypes.EpochChange
	taskCount        int64
	closed           atomic.Bool
}

func newSnapSyncManager(syncMgr common.Sync, stateVl ledger.StateLedger, chainLedger ledger.ChainLedger, rep *repo.Repo, epochMgr *epochmgr.EpochManager,
	snapHeader *types.BlockHeader, logger logrus.FieldLogger, ctx context.Context) (*snapSyncManager, error) {
	s := &snapSyncManager{
		syncMgr:     syncMgr,
		stateVl:     stateVl,
		chainLedger: chainLedger,
		rep:         rep,
		epochMgr:    epochMgr,
		logger:      logger,
		closeCh:     make(chan struct{}),
		ctx:         ctx,
	}

	// 1. load snap meta
	sm, err := s.loadSnapMeta(snapHeader, rep.P2PKeystore.P2PID(), rep.SyncArgs)
	if err != nil {
		return nil, fmt.Errorf("load snap meta: %w", err)
	}
	s.snapMeta = sm

	// 2. switch to snapshot mode
	err = s.syncMgr.SwitchMode(common.SyncModeSnapshot)
	if err != nil {
		return nil, fmt.Errorf("switch mode err: %w", err)
	}

	// 3. check whether snap sync is legal
	if err = s.checkSnapSync(); err != nil {
		return nil, fmt.Errorf("check snap sync err: %w", err)
	}
	return s, nil
}

func (s *snapSyncManager) productTask(num int64) {
	s.taskCount = num
}

func (s *snapSyncManager) consumeTask() {
	s.taskCount--
	if s.taskCount == 0 {
		close(s.closeCh)
	}
}

func (s *snapSyncManager) addEpochState(qckpt types.QuorumCheckpoint) {
	s.epochChangeCache = append(s.epochChangeCache, &consensustypes.EpochChange{
		QuorumCheckpoint: qckpt,
	})
}

func (s *snapSyncManager) clearEpochChangeCache() {
	s.epochChangeCache = make([]*consensustypes.EpochChange, 0)
}

func (s *snapSyncManager) beyondEpochChangeSize() bool {
	return len(s.epochChangeCache) > int(s.rep.Config.Sync.MaxEpochSize)
}

func (s *snapSyncManager) getLastEpochChange() (*consensustypes.EpochChange, error) {
	return lo.Last(s.epochChangeCache)
}

// Start starts snap sync process, asynchronously do the following tasks:
// 1. generate snapshot and save it to snapshot storage.
// 2. verify whether trie snapshot is legal.
// 3. sync epochState. after sync all epochChanges, start sync CommitData:
// 3.1 if epochChanges too large, split into multiple sync CommitData tasks.
func (s *snapSyncManager) Start() error {
	taskNum := int64(3)
	s.productTask(taskNum)

	// 1. start sync epoch state task
	epochStateRecvCh := make(chan containers.Tuple[types.QuorumCheckpoint, error, bool], repo.ChannelSize)
	syncEpochStateTime := time.Now()
	err := s.startSyncEpochStateTask(epochStateRecvCh)
	if err != nil {
		return fmt.Errorf("start sync epoch state task: %w", err)
	}

	// 2. start generate snapshot store task
	genSnapStoreTaskCh := make(chan error, 1)
	s.startGenerateSnapShotStoreTask(genSnapStoreTaskCh)

	// 3. start verify state trie task
	verifyTrieTime := time.Now()
	verifiedCh := s.startVerifyStateTrieTask()

	for {
		select {
		case <-s.ctx.Done():
			s.stop()
		case <-s.closeCh:
			s.logger.Info("stop snap sync")
			if s.taskCount == 0 {
				return nil
			} else {
				return fmt.Errorf("snap sync tasks exited abnormally！！！")
			}
		case resErr := <-genSnapStoreTaskCh:
			if resErr != nil {
				s.logger.Errorf("generate snapshot store failed, err: %s", resErr)
				if !s.closed.Load() {
					close(s.closeCh)
					s.closed.Store(true)
				}
				continue
			} else {
				s.logger.WithFields(logrus.Fields{
					"target height": s.snapMeta.snapBlockHeader.Number,
				}).Info("generate snapshot store success")
				s.consumeTask()
			}
		case res := <-epochStateRecvCh:
			state, err, taskDone := res.Unpack()
			if err != nil {
				s.logger.Errorf("unpack epoch state failed, err: %s", err)
				if !s.closed.Load() {
					close(s.closeCh)
					s.closed.Store(true)
				}
				continue
			} else {
				notifySyncCommitData := func() bool {
					return taskDone || len(s.epochChangeCache) >= int(s.rep.Config.Sync.MaxEpochSize)
				}
				// 1. store epoch state if exist
				if state != nil {
					if err = s.epochMgr.StoreEpochState(state); err != nil {
						s.logger.Errorf("store epoch state failed, err: %s", err)
						if !s.closed.Load() {
							close(s.closeCh)
							s.closed.Store(true)
						}
						continue
					}

					// 2. fill epoch changes for sync params
					s.addEpochState(state)
				}

				// 3. check if start sync commitData task
				if notifySyncCommitData() {
					// 3.1 start sync commitData task
					s.productTask(1)
					syncTaskRecvCh := make(chan containers.Tuple[uint64, string, error], 1)
					if err = s.startSyncCommitDataTask(syncTaskRecvCh); err != nil {
						return err
					}

					// 3.2 wait sync commitData task done, to avoid slow sync commitData,
					// which could cause the epochStateCache to occupy a large amount of memory.
					select {
					case <-s.closeCh:
						return nil
					case d := <-syncTaskRecvCh:
						targetHeight, latestHash, resErr := d.Unpack()
						if resErr != nil {
							s.logger.Errorf("start sync commitData task failed, err: %s", resErr)
							if !s.closed.Load() {
								close(s.closeCh)
								s.closed.Store(true)
							}
							continue
						}

						// 3.3 update snapMeta and consume task, start next sync epoch state task
						s.snapMeta.startHeight = targetHeight + 1
						s.snapMeta.latestBlockHash = latestHash
						s.clearEpochChangeCache()
						s.consumeTask()
					}
				}

				if taskDone {
					if state != nil {
						s.logger.WithFields(logrus.Fields{
							"cost":         time.Since(syncEpochStateTime),
							"target epoch": state.NextEpoch(),
						}).Info("sync commitData and persist epoch state success")
					}
					s.consumeTask()
				}
			}

		case valid := <-verifiedCh:
			if valid {
				s.logger.WithFields(logrus.Fields{
					"cost":          time.Since(verifyTrieTime),
					"target height": s.snapMeta.snapBlockHeader.Number,
				}).Info("verify state trie success")
				s.consumeTask()
			} else {
				s.logger.WithFields(logrus.Fields{
					"target height": s.snapMeta.snapBlockHeader.Number,
				}).Error("verify state trie failed")
				if !s.closed.Load() {
					close(s.closeCh)
					s.closed.Store(true)
				}
				continue
			}
		}
	}
}

func (s *snapSyncManager) stop() {
	if !s.closed.Load() {
		close(s.closeCh)
		s.closed.Store(true)
	}
}

// prepareSnapSyncMeta prepares snapshot syncing meta, including some phases:
// 1. open state storage
// 2. open snapshot storage(should be empty)
// 3. fill snapshot meta in axm
//   - block header: the target block header of state ledger matched
//   - snapPeers: the remote peers of snapshot syncing
//   - epochInfo: the target epoch info of state ledger matched

func (s *snapSyncManager) loadSnapMeta(header *types.BlockHeader, selfPeerId string, args *repo.SyncArgs) (*snapMeta, error) {
	meta := s.chainLedger.GetChainMeta()
	syncStartHeight := meta.Height + 1
	epochContract := framework.EpochManagerBuildConfig.Build(syscommon.NewViewVMContext(s.stateVl))
	currentEpochInfo, err := epochContract.CurrentEpoch()
	if err != nil {
		return nil, fmt.Errorf("get current epoch info: %w", err)
	}

	snapPersistedEpoch := currentEpochInfo.Epoch - 1
	if currentEpochInfo.EpochPeriod+currentEpochInfo.StartBlock-1 == header.Number {
		snapPersistedEpoch = currentEpochInfo.Epoch
	}

	// if local node is started with specified nodes for synchronization,
	// the specified nodes will be used instead of the snapshot meta
	nodeContract := framework.NodeManagerBuildConfig.Build(syscommon.NewViewVMContext(s.stateVl))
	nodeInfos, _, err := nodeContract.GetActiveValidatorSet()
	if err != nil {
		return nil, fmt.Errorf("get node info: %w", err)
	}

	p := lo.FilterMap(nodeInfos, func(info node_manager.NodeInfo, _ int) (*rbft.QuorumValidator, bool) {
		if info.P2PID == selfPeerId {
			return nil, false
		}
		return &rbft.QuorumValidator{
			Id:     info.ID,
			PeerId: info.P2PID,
		}, true
	})
	rawPeers := &rbft.QuorumValidators{Validators: p}

	if args != nil && args.RemotePeers != nil && len(args.RemotePeers.Validators) != 0 {
		rawPeers = args.RemotePeers
	}
	// flatten peers
	peers := lo.FlatMap(rawPeers.Validators, func(p *rbft.QuorumValidator, _ int) []*common.Node {
		return []*common.Node{
			{
				Id:     p.Id,
				PeerID: p.PeerId,
			},
		}
	})

	return &snapMeta{
		snapBlockHeader:  header,
		snapPeers:        peers,
		snapPersistEpoch: snapPersistedEpoch,
		startHeight:      syncStartHeight,
		latestBlockHash:  meta.BlockHash.String(),
	}, nil
}

func (s *snapSyncManager) checkSnapSync() error {
	latestHeight := s.snapMeta.startHeight - 1
	if latestHeight != s.rep.GenesisConfig.EpochInfo.StartBlock {
		return fmt.Errorf("invalid latest height: %d, chain ledger should be empty in snap sync mode", latestHeight)
	}
	return nil
}

func (s *snapSyncManager) startSyncEpochStateTask(epochStateSendCh chan<- containers.Tuple[types.QuorumCheckpoint, error, bool]) error {
	startEpcNum := s.rep.GenesisConfig.EpochInfo.Epoch
	// 2. fill snap sync config option
	opts := []common.Option{
		common.WithPeers(s.snapMeta.snapPeers),
		common.WithStartEpochChangeNum(startEpcNum),
		common.WithLatestPersistEpoch(s.epochMgr.GetLatestEpoch()),
		common.WithSnapCurrentEpoch(s.snapMeta.snapPersistEpoch),
		common.WithEpochChangeSendCh(epochStateSendCh),
	}

	// 3. prepare snap sync info
	err := s.syncMgr.Prepare(opts...)
	if err != nil {
		return fmt.Errorf("prepare sync err: %w", err)
	}
	return nil
}

func (s *snapSyncManager) startGenerateSnapShotStoreTask(tackDoneSendCh chan error) {
	go s.stateVl.GenerateSnapshot(s.snapMeta.snapBlockHeader, tackDoneSendCh)
}

func (s *snapSyncManager) startVerifyStateTrieTask() <-chan bool {
	resultCh := make(chan bool, 1)
	go func(resCh chan bool) {
		res, err := s.stateVl.VerifyTrie(s.snapMeta.snapBlockHeader)
		if err != nil {
			resCh <- false
		}
		resCh <- res
	}(resultCh)
	return resultCh
}

func (s *snapSyncManager) startSyncCommitDataTask(taskDoneSendCh chan<- containers.Tuple[uint64, string, error]) error {
	syncTaskResCh := make(chan error, 1)
	params, err := s.genSnapSyncParams()
	if err != nil {
		return err
	}
	if err = s.syncMgr.StartSync(params, syncTaskResCh); err != nil {
		return err
	}

	go func(syncTaskDoneCh chan error) {
		for {
			select {
			case <-s.closeCh:
				return
			case err = <-syncTaskDoneCh:
				if err != nil {
					s.logger.WithError(err).Error("sync commit data task failed")
					channel.SafeSend(taskDoneSendCh, containers.Pack3(params.CurHeight, params.LatestBlockHash, err), s.closeCh)
					return
				}
			case data := <-s.syncMgr.Commit():
				now := time.Now()
				cd, ok := data.(*common.SnapCommitData)
				if !ok {
					panic("invalid commit data type")
				}
				if err = s.persistChainData(cd); err != nil {
					s.logger.WithError(err).Error("persist chain data failed")
					channel.SafeSend(taskDoneSendCh, containers.Pack3(cd.TargetHeight, params.LatestBlockHash, err), s.closeCh)
					return
				}
				s.logger.WithFields(logrus.Fields{
					"Height": cd.TargetHeight,
					"target": params.TargetHeight,
					"cost":   time.Since(now),
				}).Info("persist chain data task")

				if cd.TargetHeight == params.TargetHeight {
					channel.SafeSend(taskDoneSendCh, containers.Pack3(cd.TargetHeight, cd.TargetHash, err), s.closeCh)
					return
				}
			}
		}
	}(syncTaskResCh)
	return nil
}

func (s *snapSyncManager) genSnapSyncParams() (*common.SyncParams, error) {
	targetHeight := s.snapMeta.snapBlockHeader.Number
	blockHash := s.snapMeta.snapBlockHeader.Hash().String()
	if s.beyondEpochChangeSize() {
		lastEpochState, err := s.getLastEpochChange()
		if err != nil {
			return nil, err
		}
		targetHeight = lastEpochState.GetHeight()
		blockHash = lastEpochState.GetStateDigest()
	}
	var ckpt types.QuorumCheckpoint
	switch s.rep.Config.Consensus.Type {
	case repo.ConsensusTypeDagBft:
		ckpt = &consensustypes.DagbftQuorumCheckpoint{
			QuorumCheckpoint: &dagtypes.QuorumCheckpoint{
				QuorumCheckpoint: protos.QuorumCheckpoint{
					Checkpoint: &protos.Checkpoint{
						ExecuteState: &protos.ExecuteState{
							Height:    targetHeight,
							StateRoot: blockHash,
						},
					},
				},
			},
		}
	case repo.ConsensusTypeRbft:
		ckpt = &rbft.RbftQuorumCheckpoint{
			QuorumCheckpoint: &rbft.QuorumCheckpoint{
				Checkpoint: &rbft.Checkpoint{
					ExecuteState: &rbft.Checkpoint_ExecuteState{
						Height: targetHeight,
						Digest: blockHash,
					},
				},
			},
		}
	}

	epc := s.epochChangeCache
	latestBlockHash := s.snapMeta.latestBlockHash
	startHeight := s.snapMeta.startHeight
	return &common.SyncParams{
		Peers:            s.snapMeta.snapPeers,
		LatestBlockHash:  latestBlockHash,
		Quorum:           consensuscommon.GetQuorum(s.rep.Config.Consensus.Type, len(s.snapMeta.snapPeers)),
		CurHeight:        startHeight,
		TargetHeight:     targetHeight,
		QuorumCheckpoint: ckpt,
		EpochChanges:     epc,
	}, nil
}

func (s *snapSyncManager) persistChainData(data *common.SnapCommitData) error {
	var batchBlock []*types.Block
	var batchReceipts [][]*types.Receipt
	for _, commitData := range data.Data {
		chainData, ok := commitData.(*common.ChainData)
		if !ok {
			return fmt.Errorf("invalid commit data type: %T", commitData)
		}
		batchBlock = append(batchBlock, chainData.Block)
		batchReceipts = append(batchReceipts, chainData.Receipts)
	}
	if len(batchBlock) > 0 {
		if err := s.chainLedger.BatchPersistExecutionResult(batchBlock, batchReceipts); err != nil {
			return err
		}
	}
	return nil
}
