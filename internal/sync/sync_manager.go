package sync

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/internal/sync/full_sync"
	"github.com/axiomesh/axiom-ledger/internal/sync/snap_sync"
	network2 "github.com/axiomesh/axiom-p2p"
	"github.com/gammazero/workerpool"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var _ common.Sync = (*SyncManager)(nil)

type SyncManager struct {
	mode               common.SyncMode
	modeConstructor    common.ISyncConstructor
	conf               repo.Sync
	syncStatus         atomic.Bool         // sync status
	initPeers          []*common.Peer      // p2p set of latest epoch validatorSet with init
	peers              []*common.Peer      // p2p set of latest epoch validatorSet
	quorum             uint64              // quorum of latest epoch validatorSet
	curHeight          uint64              // current commitDataCache which we need sync
	targetHeight       uint64              // sync target commitDataCache height
	recvBlockSize      atomic.Int64        // current chunk had received commitDataCache size
	latestCheckedState *pb.CheckpointState // latest checked commitDataCache state
	requesters         sync.Map            // requester map
	requesterLen       atomic.Int64        // requester length

	quorumCheckpoint  *consensus.SignedCheckpoint // latest checkpoint from remote
	epochChanges      []*consensus.EpochChange    // every epoch change which the node behind
	getBlockFunc      func(height uint64) (*types.Block, error)
	getReceiptsFunc   func(height uint64) ([]*types.Receipt, error)
	getEpochStateFunc func(key []byte) []byte

	network               network.Network
	chainDataRequestPipe  network.Pipe
	chainDataResponsePipe network.Pipe
	blockRequestPipe      network.Pipe
	blockDataResponsePipe network.Pipe

	commitDataCache []common.CommitData // store commitDataCache of a chunk temporary

	chunk            *common.Chunk                 // every chunk task
	recvStateCh      chan *common.WrapperStateResp // receive state from remote peer
	quitStateCh      chan error                    // quit state channel
	stateTaskDone    atomic.Bool                   // state task done signal
	requesterCh      chan struct{}                 // chunk task done signal
	validChunkTaskCh chan struct{}                 // start validate chunk task signal

	invalidRequestCh chan *common.InvalidMsg // timeout or invalid of sync Block request

	ctx    context.Context
	cancel context.CancelFunc

	syncCtx    context.Context
	syncCancel context.CancelFunc

	logger logrus.FieldLogger
}

func NewSyncManager(logger logrus.FieldLogger, fn func(height uint64) (*types.Block, error),
	receiptFn func(height uint64) ([]*types.Receipt, error), getEpochStateFunc func(key []byte) []byte, network network.Network, cnf repo.Sync) (*SyncManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	syncMgr := &SyncManager{
		mode:              common.SyncModeFull,
		modeConstructor:   newModeConstructor(common.SyncModeFull, common.WithContext(ctx)),
		logger:            logger,
		invalidRequestCh:  make(chan *common.InvalidMsg, cnf.ConcurrencyLimit),
		recvStateCh:       make(chan *common.WrapperStateResp, cnf.ConcurrencyLimit),
		validChunkTaskCh:  make(chan struct{}, 1),
		quitStateCh:       make(chan error, 1),
		requesterCh:       make(chan struct{}, cnf.ConcurrencyLimit),
		getBlockFunc:      fn,
		getReceiptsFunc:   receiptFn,
		getEpochStateFunc: getEpochStateFunc,
		network:           network,
		conf:              cnf,

		ctx:    ctx,
		cancel: cancel,
	}

	// init sync block pipe
	reqPipe, err := syncMgr.network.CreatePipe(syncMgr.ctx, common.SyncBlockRequestPipe)
	if err != nil {
		return nil, err
	}
	syncMgr.blockRequestPipe = reqPipe

	respPipe, err := syncMgr.network.CreatePipe(syncMgr.ctx, common.SyncBlockResponsePipe)
	if err != nil {
		return nil, err
	}
	syncMgr.blockDataResponsePipe = respPipe

	// init sync chainData pipe
	chainReqPipe, err := syncMgr.network.CreatePipe(syncMgr.ctx, common.SyncChainDataRequestPipe)
	if err != nil {
		return nil, err
	}
	syncMgr.chainDataRequestPipe = chainReqPipe

	chainRespPipe, err := syncMgr.network.CreatePipe(syncMgr.ctx, common.SyncChainDataResponsePipe)
	if err != nil {
		return nil, err
	}
	syncMgr.chainDataResponsePipe = chainRespPipe

	// init syncStatus
	syncMgr.syncStatus.Store(false)

	syncMgr.logger.Info("Init sync manager success")

	return syncMgr, nil
}

func newModeConstructor(mode common.SyncMode, opt ...common.ModeOption) common.ISyncConstructor {
	conf := &common.ModeConfig{}
	for _, o := range opt {
		o(conf)
	}
	switch mode {
	case common.SyncModeFull:
		return full_sync.NewFullSync(conf.Ctx)
	case common.SyncModeSnapshot:
		return snap_sync.NewSnapSync(conf.Logger, conf.Ctx, conf.Network)
	}
	return nil
}

func (sm *SyncManager) produceRequester(count uint64) {
	for i := 0; i < int(count); i++ {
		sm.requesterCh <- struct{}{}
	}
}
func (sm *SyncManager) consumeRequester() chan struct{} {
	return sm.requesterCh
}

func (sm *SyncManager) postInvalidMsg(msg *common.InvalidMsg) {
	sm.invalidRequestCh <- msg
}

func (sm *SyncManager) processChunkTask(syncCount uint64, startTime time.Time, syncTaskDoneCh chan error) {
	invalidReqs, err := sm.validateChunk()
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Validate chunk failed")
		syncTaskDoneCh <- err
		return
	}

	if len(invalidReqs) != 0 {
		lo.ForEach(invalidReqs, func(req *common.InvalidMsg, index int) {
			sm.postInvalidMsg(req)
			sm.logger.WithFields(logrus.Fields{
				"peer":   req.NodeID,
				"height": req.Height,
				"err":    req.ErrMsg,
			}).Warning("Receive Invalid commitData")
		})
		return
	}

	lastR := sm.getRequester(sm.curHeight + sm.chunk.ChunkSize - 1)
	if lastR == nil {
		err = errors.New("get last requester failed")
		sm.logger.WithFields(logrus.Fields{
			"height": sm.curHeight + sm.chunk.ChunkSize - 1,
		}).Error(err.Error())
		syncTaskDoneCh <- err
		return
	}
	err = sm.validateChunkState(lastR.commitData.GetHeight(), lastR.commitData.GetHash())
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Validate last chunk state failed")
		invalidMsg := &common.InvalidMsg{
			NodeID: lastR.peerID,
			Height: lastR.commitData.GetHeight(),
			Typ:    common.SyncMsgType_InvalidBlock,
		}
		sm.postInvalidMsg(invalidMsg)
		return
	}

	// release requester and send commitData to commitDataCh
	sm.requesters.Range(func(height, r any) bool {
		if r.(*requester).commitData == nil {
			err = errors.New("requester commitData is nil")
			sm.logger.WithFields(logrus.Fields{
				"height": sm.curHeight + sm.chunk.ChunkSize - 1,
			}).Error(err.Error())
			syncTaskDoneCh <- err
			return false
		}

		sm.commitDataCache[height.(uint64)-sm.curHeight] = r.(*requester).commitData
		return true
	})

	// if commitDataCache is not full, continue to receive commitDataCache
	if len(sm.commitDataCache) != int(sm.chunk.ChunkSize) {
		return
	}

	blocksLen := len(sm.commitDataCache)
	sm.updateLatestCheckedState(sm.commitDataCache[blocksLen-1].GetHeight(), sm.commitDataCache[blocksLen-1].GetHash())

	// if valid chunk task done, release all requester
	lo.ForEach(sm.commitDataCache, func(commitData common.CommitData, index int) {
		sm.releaseRequester(commitData.GetHeight())
	})

	if sm.chunk.CheckPoint != nil {
		idx := int(sm.chunk.CheckPoint.Height - sm.curHeight)
		if idx < 0 || idx > blocksLen-1 {
			sm.logger.Errorf("chunk checkpoint index out of range, checkpoint height:%d, current Height:%d, "+
				"commitData len:%d", sm.chunk.CheckPoint.Height, sm.curHeight, blocksLen)
			return
		}

		// if checkpoint is not equal to last commitDataCache, it means we sync wrong commitDataCache, panic it
		if err = sm.verifyChunkCheckpoint(sm.commitDataCache[idx]); err != nil {
			sm.logger.Errorf("verify chunk checkpoint failed: %s", err)
			syncTaskDoneCh <- err
		}
	}

	sm.modeConstructor.PostCommitData(sm.commitDataCache)

	if sm.curHeight+sm.chunk.ChunkSize-1 == sm.targetHeight {
		if err = sm.stopSync(); err != nil {
			sm.logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Stop sync failed")
		}
		sm.logger.WithFields(logrus.Fields{
			"count":  syncCount,
			"target": sm.targetHeight,
			"elapse": time.Since(startTime).Seconds(),
		}).Info("Block sync done")
		blockSyncDuration.WithLabelValues(strconv.Itoa(int(syncCount))).Observe(time.Since(startTime).Seconds())
		syncTaskDoneCh <- nil
	} else {
		sm.logger.WithFields(logrus.Fields{
			"start":  sm.curHeight,
			"target": sm.curHeight + sm.chunk.ChunkSize - 1,
		}).Info("chunk task has done")
		sm.updateStatus()
		// produce many new requesters for new chunk
		sm.produceRequester(sm.chunk.ChunkSize)
	}
}

func (sm *SyncManager) StartSync(params *common.SyncParams, syncTaskDoneCh chan error) error {
	now := time.Now()
	syncCount := params.TargetHeight - params.CurHeight + 1
	syncCtx, syncCancel := context.WithCancel(context.Background())
	sm.syncCtx = syncCtx
	sm.syncCancel = syncCancel

	// 1. update commitDataCache sync info
	sm.InitBlockSyncInfo(params.Peers, params.LatestBlockHash, params.Quorum, params.CurHeight, params.TargetHeight, params.QuorumCheckpoint, params.EpochChanges...)

	// 2. send sync state request to all validators, waiting for Quorum response
	if sm.curHeight != 1 {
		err := sm.requestSyncState(sm.curHeight-1, params.LatestBlockHash)
		if err != nil {
			syncTaskDoneCh <- err
			return err
		}

		sm.logger.WithFields(logrus.Fields{
			"Quorum": sm.quorum,
		}).Info("Receive Quorum response")
	}

	// 3. switch sync status to true, if switch failed, return error
	if err := sm.switchSyncStatus(true); err != nil {
		syncTaskDoneCh <- err
		return err
	}

	// 4. start listen sync commitDataCache response
	go sm.listenSyncCommitDataResponse()

	// 5. produce requesters for first chunk
	sm.produceRequester(sm.chunk.ChunkSize)

	// 6. start sync commitDataCache task
	go func() {
		for {
			select {
			case <-sm.syncCtx.Done():
				return
			case msg := <-sm.invalidRequestCh:
				sm.handleInvalidRequest(msg)
			case <-sm.validChunkTaskCh:
				sm.processChunkTask(syncCount, now, syncTaskDoneCh)
			case <-sm.consumeRequester():
				sm.makeRequesters(sm.curHeight + uint64(sm.requesterLen.Load()))
			}
		}
	}()

	return nil
}

func (sm *SyncManager) validateChunkState(localHeight uint64, localHash string) error {
	return sm.requestSyncState(localHeight, localHash)
}

func (sm *SyncManager) validateChunk() ([]*common.InvalidMsg, error) {
	parentHash := sm.latestCheckedState.Digest
	for i := sm.curHeight; i < sm.curHeight+sm.chunk.ChunkSize; i++ {
		r := sm.getRequester(i)
		if r == nil {
			return nil, fmt.Errorf("requester[height:%d] is nil", i)
		}
		cd := r.commitData
		if cd == nil {
			sm.logger.WithFields(logrus.Fields{
				"height": i,
			}).Error("Block is nil")
			return []*common.InvalidMsg{
				{
					NodeID: r.peerID,
					Height: i,
					Typ:    common.SyncMsgType_TimeoutBlock,
				},
			}, nil
		}

		if cd.GetParentHash() != parentHash {
			sm.logger.WithFields(logrus.Fields{
				"height":               i,
				"expect parent hash":   parentHash,
				"expect parent height": i - 1,
				"actual parent hash":   cd.GetParentHash(),
			}).Error("Block parent hash is not equal to latest checked state")

			invalidMsgs := make([]*common.InvalidMsg, 0)

			// if we have not previous requester, it means we had already checked previous commitDataCache,
			// so we just return current invalid commitDataCache
			prevR := sm.getRequester(i - 1)
			if prevR != nil {
				// we are not sure which commitDataCache is wrong,
				// maybe parent commitDataCache has wrong hash, maybe current commitDataCache has wrong parent hash, so we return two invalidMsg
				prevInvalidMsg := &common.InvalidMsg{
					NodeID: prevR.peerID,
					Height: i - 1,
					Typ:    common.SyncMsgType_InvalidBlock,
				}
				invalidMsgs = append(invalidMsgs, prevInvalidMsg)
			}
			invalidMsgs = append(invalidMsgs, &common.InvalidMsg{
				NodeID: r.peerID,
				Height: i,
				Typ:    common.SyncMsgType_InvalidBlock,
			})

			return invalidMsgs, nil
		}

		parentHash = cd.GetHash()
	}
	return nil, nil
}

func (sm *SyncManager) updateLatestCheckedState(height uint64, digest string) {
	sm.latestCheckedState = &pb.CheckpointState{
		Height: height,
		Digest: digest,
	}
}

func (sm *SyncManager) InitBlockSyncInfo(peers []string, latestBlockHash string, quorum, curHeight, targetHeight uint64,
	quorumCheckpoint *consensus.SignedCheckpoint, epc ...*consensus.EpochChange) {
	sm.peers = make([]*common.Peer, len(peers))
	sm.initPeers = make([]*common.Peer, len(peers))
	lo.ForEach(peers, func(p string, index int) {
		sm.peers[index] = &common.Peer{
			PeerID:       p,
			TimeoutCount: 0,
		}
	})
	copy(sm.initPeers, sm.peers)
	sm.quorum = quorum
	sm.curHeight = curHeight
	sm.targetHeight = targetHeight
	sm.quorumCheckpoint = quorumCheckpoint
	sm.epochChanges = epc
	sm.recvBlockSize.Store(0)

	sm.updateLatestCheckedState(curHeight-1, latestBlockHash)

	// init chunk
	sm.initChunk()

	sm.commitDataCache = make([]common.CommitData, sm.chunk.ChunkSize)
}

func (sm *SyncManager) initChunk() {
	chunkSize := sm.targetHeight - sm.curHeight + 1
	if chunkSize > sm.conf.ConcurrencyLimit {
		chunkSize = sm.conf.ConcurrencyLimit
	}

	// if we have epoch change, chunk size need smaller than epoch size
	if len(sm.epochChanges) != 0 {
		latestEpoch := sm.epochChanges[0]
		if latestEpoch != nil {
			epochSize := latestEpoch.GetCheckpoint().Checkpoint.Height() - sm.curHeight + 1
			if epochSize < chunkSize {
				chunkSize = epochSize
			}
		}
	}

	sm.chunk = &common.Chunk{
		ChunkSize: chunkSize,
	}

	chunkMaxHeight := sm.curHeight + chunkSize - 1

	var chunkCheckpoint *pb.CheckpointState

	if len(sm.epochChanges) != 0 {
		chunkCheckpoint = &pb.CheckpointState{
			Height: sm.epochChanges[0].GetCheckpoint().Checkpoint.Height(),
			Digest: sm.epochChanges[0].GetCheckpoint().Checkpoint.Digest(),
		}
	} else {
		chunkCheckpoint = &pb.CheckpointState{
			Height: sm.quorumCheckpoint.Height(),
			Digest: sm.quorumCheckpoint.Digest(),
		}
	}
	sm.chunk.FillCheckPoint(chunkMaxHeight, chunkCheckpoint)
}

func (sm *SyncManager) switchSyncStatus(status bool) error {
	if sm.syncStatus.Load() == status {
		return fmt.Errorf("status is already %v", status)
	}
	sm.syncStatus.Store(status)
	sm.logger.Info("SwitchSyncStatus: status is ", status)
	return nil
}

func (sm *SyncManager) listenSyncStateResp(ctx context.Context, cancel context.CancelFunc, height uint64, localHash string) {
	diffState := make(map[string][]string)

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-sm.recvStateCh:
			sm.handleSyncStateResp(resp, diffState, height, localHash, cancel)
		}
	}
}

func (sm *SyncManager) requestSyncState(height uint64, localHash string) error {
	sm.logger.WithFields(logrus.Fields{
		"height": height,
	}).Info("Prepare request sync state")
	sm.stateTaskDone.Store(false)

	// 1. start listen sync state response
	stateCtx, stateCancel := context.WithCancel(context.Background())
	go sm.listenSyncStateResp(stateCtx, stateCancel, height, localHash)

	wp := workerpool.New(len(sm.peers))
	// send sync state request to all validators, check our local state(latest commitDataCache) is equal to Quorum state
	req := &pb.SyncStateRequest{
		Height: height,
	}
	data, err := req.MarshalVT()
	if err != nil {
		return err
	}

	// 2. send sync state request to all validators asynchronously
	// because Peers num maybe too small to cannot reach Quorum, so we use initPeers to send sync state
	lo.ForEach(sm.initPeers, func(p *common.Peer, index int) {
		select {
		case <-stateCtx.Done():
			wp.Stop()
			sm.logger.Debug("receive quit signal, Quit request state")
			return
		default:
			wp.Submit(func() {
				if err = retry.Retry(func(attempt uint) error {
					select {
					case <-stateCtx.Done():
						sm.logger.WithFields(logrus.Fields{
							"peer":   p.PeerID,
							"height": height,
						}).Debug("receive quit signal, Quit request state")
						return nil
					default:
						sm.logger.WithFields(logrus.Fields{
							"peer":   p.PeerID,
							"height": height,
						}).Debug("start send sync state request")
						resp, err := sm.network.Send(p.PeerID, &pb.Message{
							Type: pb.Message_SYNC_STATE_REQUEST,
							Data: data,
						})
						if err != nil {
							sm.logger.WithFields(logrus.Fields{
								"peer": p.PeerID,
								"err":  err,
							}).Warn("Send sync state request failed")
							return err
						}

						if err = sm.isValidSyncResponse(resp, p.PeerID); err != nil {
							sm.logger.WithFields(logrus.Fields{
								"peer": p.PeerID,
								"err":  err,
							}).Warn("Invalid sync state response")

							return fmt.Errorf("invalid sync state response: %s", err)
						}

						stateResp := &pb.SyncStateResponse{}
						if err = stateResp.UnmarshalVT(resp.Data); err != nil {
							return fmt.Errorf("unmarshal sync state response failed: %s", err)
						}

						select {
						case <-stateCtx.Done():
							sm.logger.WithFields(logrus.Fields{
								"peer":   p.PeerID,
								"height": height,
							}).Debug("receive quit signal, Quit request state")
							return nil
						default:
							hash := sha256.Sum256(resp.Data)
							sm.recvStateCh <- &common.WrapperStateResp{
								PeerID: resp.From,
								Hash:   types.NewHash(hash[:]).String(),
								Resp:   stateResp,
							}
						}
						return nil
					}
				}, strategy.Backoff(backoff.Fibonacci(500*time.Millisecond))); err != nil {
					sm.logger.Errorf("Retry send sync state request failed: %s", err)
					return
				}
				sm.logger.WithFields(logrus.Fields{
					"peer":   p.PeerID,
					"height": height,
				}).Debug("Send sync state request success")
			})
		}
	})

	for {
		select {
		case stateErr := <-sm.quitStateCh:
			if stateErr != nil {
				return fmt.Errorf("request state failed: %s", stateErr)
			}
			return nil
		case <-time.After(sm.conf.WaitStatesTimeout.ToDuration()):
			sm.logger.WithFields(logrus.Fields{
				"height": height,
			}).Warn("Request state timeout")

			sm.stateTaskDone.Store(true)
			return fmt.Errorf("request state timeout: height:%d", height)
		}
	}
}

func (sm *SyncManager) isValidSyncResponse(msg *pb.Message, id string) error {
	if msg == nil || msg.Data == nil {
		return errors.New("sync response is nil")
	}

	if msg.From != id {
		return fmt.Errorf("receive different peer sync response, expect peer id is %s,"+
			" but receive peer id is %s", id, msg.From)
	}

	return nil
}

func (sm *SyncManager) handleCommitDataRequest(msg *network2.PipeMsg, mode common.SyncMode) ([]byte, error) {
	var (
		requestHeight uint64
		data          []byte
		resp          *pb.Message
	)

	req := common.CommitDataRequestConstructor[mode]
	if err := req.UnmarshalVT(msg.Data); err != nil {
		sm.logger.Errorf("Unmarshal sync commitData request failed: %s", err)
		return nil, err
	}
	requestHeight = req.GetHeight()

	block, err := sm.getBlockFunc(requestHeight)
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"from": msg.From,
			"err":  err,
		}).Error("Get commitData failed")
		return nil, err
	}

	blockBytes, err := block.Marshal()
	if err != nil {
		sm.logger.Errorf("Marshal commitData failed: %s", err)
		return nil, err
	}

	switch mode {
	case common.SyncModeFull:
		commitDataResp := &pb.SyncBlockResponse{Block: blockBytes}
		commitDataBytes, err := commitDataResp.MarshalVT()
		if err != nil {
			sm.logger.Errorf("Marshal sync commitData response failed: %s", err)
			return nil, err
		}
		resp = &pb.Message{
			Type: pb.Message_SYNC_BLOCK_RESPONSE,
			Data: commitDataBytes,
		}

	case common.SyncModeSnapshot:
		commitDataResp := &pb.SyncChainDataResponse{Block: blockBytes}

		// get receipts by block height
		receipts, err := sm.getReceiptsFunc(requestHeight)
		if err != nil {
			sm.logger.WithFields(logrus.Fields{
				"from": msg.From,
				"err":  err,
			}).Error("Get receipts failed")
			return nil, err
		}
		receiptsBytes, err := types.MarshalReceipts(receipts)
		if err != nil {
			sm.logger.Errorf("Marshal receipts failed: %s", err)
			return nil, err
		}
		commitDataResp.Receipts = receiptsBytes
		commitDataBytes, err := commitDataResp.MarshalVT()
		if err != nil {
			sm.logger.Errorf("Marshal sync commitData response failed: %s", err)
			return nil, err
		}
		resp = &pb.Message{
			Type: pb.Message_SYNC_CHAIN_DATA_RESPONSE,
			Data: commitDataBytes,
		}
	}

	data, err = resp.MarshalVT()
	if err != nil {
		sm.logger.Errorf("Marshal sync commitData response failed: %s", err)
		return nil, err
	}

	return data, nil
}

func (sm *SyncManager) listenSyncChainDataRequest() {
	for {
		msg := sm.chainDataRequestPipe.Receive(sm.ctx)
		if msg == nil {
			sm.logger.Info("Stop listen sync chainData request")
			return
		}

		data, err := sm.handleCommitDataRequest(msg, common.SyncModeSnapshot)
		if err != nil {
			sm.logger.Errorf("Handle sync chainData request failed: %s", err)
			continue
		}
		if err := retry.Retry(func(attempt uint) error {
			err := sm.chainDataResponsePipe.Send(sm.ctx, msg.From, data)
			if err != nil {
				sm.logger.WithFields(logrus.Fields{
					"from": msg.From,
					"err":  err,
				}).Error("Send sync commitData response failed")
				return err
			}
			sm.logger.WithFields(logrus.Fields{
				"from": msg.From,
			}).Debug("Send sync commitData response success")
			return nil
		}, strategy.Limit(common.MaxRetryCount), strategy.Wait(500*time.Millisecond)); err != nil {
			sm.logger.Errorf("Retry send sync chainData response failed: %s", err)

			continue
		}
	}
}

func (sm *SyncManager) listenSyncBlockDataRequest() {
	for {
		msg := sm.blockRequestPipe.Receive(sm.ctx)
		if msg == nil {
			sm.logger.Info("Stop listen sync block request")
			return
		}

		data, err := sm.handleCommitDataRequest(msg, common.SyncModeFull)
		if err != nil {
			sm.logger.Errorf("Handle sync block request failed: %s", err)
			continue
		}

		if err := retry.Retry(func(attempt uint) error {
			err := sm.blockDataResponsePipe.Send(sm.ctx, msg.From, data)
			if err != nil {
				sm.logger.WithFields(logrus.Fields{
					"from": msg.From,
					"err":  err,
				}).Error("Send sync block response failed")
				return err
			}
			sm.logger.WithFields(logrus.Fields{
				"from": msg.From,
			}).Debug("Send sync block response success")
			return nil
		}, strategy.Limit(common.MaxRetryCount), strategy.Wait(500*time.Millisecond)); err != nil {
			sm.logger.Errorf("Retry send sync block response failed: %s", err)

			continue
		}
	}
}

func (sm *SyncManager) listenSyncCommitDataResponse() {
	for {
		select {
		case <-sm.syncCtx.Done():
			return
		default:
			var msg *network2.PipeMsg
			switch sm.mode {
			case common.SyncModeFull:
				msg = sm.blockDataResponsePipe.Receive(sm.syncCtx)
			case common.SyncModeSnapshot:
				msg = sm.chainDataResponsePipe.Receive(sm.syncCtx)
			}
			if msg == nil {
				return
			}

			p2pMsg := &pb.Message{}
			if err := p2pMsg.UnmarshalVT(msg.Data); err != nil {
				sm.logger.Errorf("Unmarshal sync commitData response failed: %s", err)
				continue
			}

			var commitData common.CommitData
			switch sm.mode {
			case common.SyncModeFull:
				if p2pMsg.Type != pb.Message_SYNC_BLOCK_RESPONSE {
					sm.logger.Errorf("Receive invalid sync block response type: %s", p2pMsg.Type)
					continue
				}
				resp := &pb.SyncBlockResponse{}
				if err := resp.UnmarshalVT(p2pMsg.Data); err != nil {
					sm.logger.Errorf("Unmarshal sync commitData response failed: %s", err)
					continue
				}
				block := &types.Block{}
				if err := block.Unmarshal(resp.GetBlock()); err != nil {
					sm.logger.Errorf("Unmarshal block failed: %s", err)
					continue
				}
				commitData = &common.BlockData{Block: block}

			case common.SyncModeSnapshot:
				if p2pMsg.Type != pb.Message_SYNC_CHAIN_DATA_RESPONSE {
					sm.logger.Errorf("Receive invalid sync chainData response type: %s", p2pMsg.Type)
					continue
				}
				resp := &pb.SyncChainDataResponse{}
				if err := resp.UnmarshalVT(p2pMsg.Data); err != nil {
					sm.logger.Errorf("Unmarshal sync commitData response failed: %s", err)
					continue
				}
				chainData := &common.ChainData{}
				block := &types.Block{}
				if err := block.Unmarshal(resp.GetBlock()); err != nil {
					sm.logger.Errorf("Unmarshal block failed: %s", err)
					continue
				}
				receipts, err := types.UnmarshalReceipts(resp.GetReceipts())
				if err != nil {
					sm.logger.Errorf("Unmarshal receipt failed: %s", err)
					continue
				}
				chainData.Block = block
				chainData.Receipts = receipts
				commitData = chainData
			}

			err, updated := sm.addCommitData(commitData, msg.From)
			if err != nil {
				sm.logger.WithFields(logrus.Fields{
					"from": msg.From,
					"err":  err,
				}).Error("Add commitData failed")
				continue
			}

			if updated {
				if sm.collectChunkTaskDone() {
					sm.logger.WithFields(logrus.Fields{
						"latest commitDataCache": commitData.GetHeight(),
						"hash":                   commitData.GetHash(),
						"peer":                   msg.From,
					}).Debug("Receive chunk commitDataCache success")
					// send valid chunk task signal
					sm.validChunkTaskCh <- struct{}{}
				}
			}
		}
	}
}

func (sm *SyncManager) verifyChunkCheckpoint(checkCommitData common.CommitData) error {
	if sm.chunk.CheckPoint.Digest != checkCommitData.GetHash() {
		return fmt.Errorf("quorum checkpoint is not equal to current hash:[height:%d quorum hash:%s, current hash:%s]",
			sm.chunk.CheckPoint.Height, sm.chunk.CheckPoint.Digest, checkCommitData.GetHash())
	}
	return nil
}

func (sm *SyncManager) addCommitData(commitData common.CommitData, from string) (error, bool) {
	req := sm.getRequester(commitData.GetHeight())
	if req == nil {
		return fmt.Errorf("requester[height:%d] is nil", commitData.GetHeight()), false
	}

	if req.peerID != from {
		sm.logger.Warningf("receive commitData which not distribute requester, height:%d, "+
			"receive from:%s, expect from:%s, we will ignore this commitDataCache", commitData.GetHeight(), from, req.peerID)
		return nil, false
	}

	updated := false
	if req.commitData == nil {
		req.setCommitData(commitData)
		sm.increaseBlockSize()
		updated = true
	}
	sm.logger.WithFields(logrus.Fields{
		"height":         commitData.GetHeight(),
		"from":           from,
		"add_commitData": commitData.GetHash(),
		"hash":           req.commitData.GetHash(),
		"size":           sm.recvBlockSize.Load(),
	}).Debug("Receive commitData success")
	return nil, updated
}

func (sm *SyncManager) collectChunkTaskDone() bool {
	if sm.chunk.ChunkSize == 0 {
		return true
	}

	return sm.recvBlockSize.Load() >= int64(sm.chunk.ChunkSize)
}

func (sm *SyncManager) handleSyncStateResp(msg *common.WrapperStateResp, diffState map[string][]string, localHeight uint64,
	localHash string, cancel context.CancelFunc) {
	sm.logger.WithFields(logrus.Fields{
		"peer":   msg.PeerID,
		"height": msg.Resp.CheckpointState.Height,
		"digest": msg.Resp.CheckpointState.Digest,
	}).Debug("Receive sync state response")

	if sm.stateTaskDone.Load() || localHeight != msg.Resp.CheckpointState.Height {
		sm.logger.WithFields(logrus.Fields{
			"peer":   msg.PeerID,
			"height": msg.Resp.CheckpointState.Height,
			"digest": msg.Resp.CheckpointState.Digest,
		}).Debug("Receive state response after state task done, we ignore it")
		return
	}
	diffState[msg.Hash] = append(diffState[msg.Hash], msg.PeerID)

	// if Quorum state is enough, update Quorum state
	if len(diffState[msg.Hash]) >= int(sm.quorum) {
		defer cancel()
		// verify Quorum state failed, we will quit state with false
		if msg.Resp.CheckpointState.Digest != localHash {
			sm.quitState(fmt.Errorf("quorum state is not equal to current state:[height:%d quorum hash:%s, current hash:%s]",
				msg.Resp.CheckpointState.Height, msg.Resp.CheckpointState.Digest, localHash))
			return
		}

		delete(diffState, msg.Hash)
		// remove Peers which not in Quorum state
		if len(diffState) != 0 {
			wrongPeers := lo.Values(diffState)
			lo.ForEach(lo.Flatten(wrongPeers), func(peer string, _ int) {
				if empty := sm.removePeer(peer); empty {
					sm.logger.Warning("available peer is empty, will reset the Peers")
					sm.resetPeers()
				}
				sm.logger.WithFields(logrus.Fields{
					"peer":   peer,
					"height": msg.Resp.CheckpointState.Height,
				}).Warn("Remove peer which not in Quorum state")
			})
		}

		// For example, if the validator set is 4, and the Quorum is 3:
		// 1. if the current node is forked,
		// 2. validator need send state which obtaining low watermark height commitDataCache,
		// 3. validator have different low watermark height commitDataCache due to network latency,
		// 4. it can lead to state inconsistency, and the node will be stuck in the state sync process.
		sm.logger.Debug("Receive Quorum state from Peers")
		sm.quitState(nil)
	}
}

func (sm *SyncManager) status() bool {
	return sm.syncStatus.Load()
}

func (sm *SyncManager) SwitchMode(mode common.SyncMode) error {
	if sm.mode == mode {
		return fmt.Errorf("current mode is same as switch mode: %s", common.SyncModeMap[sm.mode])
	}
	if sm.status() {
		return fmt.Errorf("switch mode failed, sync status is %v", sm.status())
	}

	var (
		constructor common.ISyncConstructor
	)
	switch mode {
	case common.SyncModeFull:
		constructor = newModeConstructor(common.SyncModeFull, common.WithContext(sm.ctx))
		_, _ = constructor.Prepare(nil)
	case common.SyncModeSnapshot:
		constructor = newModeConstructor(common.SyncModeSnapshot, common.WithContext(sm.ctx), common.WithLogger(sm.logger), common.WithNetwork(sm.network))
	}
	sm.mode = mode
	sm.modeConstructor = constructor
	return nil
}

func (sm *SyncManager) Prepare(opts ...common.Option) (*common.PrepareData, error) {
	// register message handler
	err := sm.network.RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, sm.handleSyncState)
	if err != nil {
		return nil, err
	}
	// for snap sync
	err = sm.network.RegisterMsgHandler(pb.Message_FETCH_EPOCH_STATE_REQUEST, sm.handleFetchEpochState)
	if err != nil {
		return nil, err
	}
	// start handle sync block request in full mode
	go sm.listenSyncBlockDataRequest()

	// start handle sync chain data request in snap mode
	go sm.listenSyncChainDataRequest()
	sm.logger.Info("Prepare listen sync request")

	conf := &common.Config{}
	for _, opt := range opts {
		opt(conf)
	}
	return sm.modeConstructor.Prepare(conf)
}

func (sm *SyncManager) makeRequesters(height uint64) {
	peerID := sm.pickPeer(height)
	var pipe network.Pipe
	switch sm.mode {
	case common.SyncModeFull:
		pipe = sm.blockRequestPipe
	case common.SyncModeSnapshot:
		pipe = sm.chainDataRequestPipe
	}
	request := newRequester(sm.mode, sm.ctx, peerID, height, sm.invalidRequestCh, pipe)
	sm.increaseRequester(request, height)
	request.start(sm.conf.RequesterRetryTimeout.ToDuration())
}

// todo: add metrics
func (sm *SyncManager) increaseRequester(r *requester, height uint64) {
	oldR, loaded := sm.requesters.LoadOrStore(height, r)
	if !loaded {
		sm.requesterLen.Add(1)
		requesterNumber.Inc()
	} else {
		sm.logger.WithFields(logrus.Fields{
			"height": height,
		}).Warn("Make requester Error, requester is not nil, we will reset the old requester")
		oldR.(*requester).quitCh <- struct{}{}
	}
}

func (sm *SyncManager) updateStatus() {
	sm.curHeight += sm.chunk.ChunkSize
	if len(sm.epochChanges) != 0 {
		sm.epochChanges = sm.epochChanges[1:]
	}
	sm.initChunk()
	sm.resetBlockSize()

	sm.commitDataCache = make([]common.CommitData, sm.chunk.ChunkSize)
}

// todo: add metrics
func (sm *SyncManager) increaseBlockSize() {
	sm.recvBlockSize.Add(1)
	recvBlockNumber.WithLabelValues(strconv.Itoa(int(sm.chunk.ChunkSize))).Inc()
}

func (sm *SyncManager) decreaseBlockSize() {
	sm.recvBlockSize.Add(-1)
	recvBlockNumber.WithLabelValues(strconv.Itoa(int(sm.chunk.ChunkSize))).Dec()
}

func (sm *SyncManager) resetBlockSize() {
	sm.recvBlockSize.Store(0)
	sm.logger.WithFields(logrus.Fields{
		"blockSize": sm.recvBlockSize.Load(),
	}).Debug("Reset commitDataCache size")
	recvBlockNumber.WithLabelValues(strconv.Itoa(int(sm.chunk.ChunkSize))).Set(0)
}

func (sm *SyncManager) quitState(err error) {
	sm.quitStateCh <- err
	sm.stateTaskDone.Store(true)
}

func (sm *SyncManager) releaseRequester(height uint64) {
	r, loaded := sm.requesters.LoadAndDelete(height)
	if !loaded {
		sm.logger.WithFields(logrus.Fields{
			"height": height,
		}).Warn("Release requester Error, requester is nil")
	} else {
		sm.requesterLen.Add(-1)
		requesterNumber.Dec()
	}
	r.(*requester).quitCh <- struct{}{}
}

func (sm *SyncManager) handleInvalidRequest(msg *common.InvalidMsg) {
	// retry request
	r := sm.getRequester(msg.Height)
	if r == nil {
		sm.logger.Errorf("Retry request commitDataCache Error, requester[height:%d] is nil", msg.Height)
		return
	}
	switch msg.Typ {
	case common.SyncMsgType_ErrorMsg:
		sm.logger.WithFields(logrus.Fields{
			"height": msg.Height,
			"peer":   msg.NodeID,
			"err":    msg.ErrMsg,
		}).Warn("Handle error msg Block")

		invalidBlockNumber.WithLabelValues("send_request_err").Inc()
		newPeer, err := sm.pickRandomPeer(msg.NodeID)
		if err != nil {
			panic(err)
		}

		r.retryCh <- newPeer

	case common.SyncMsgType_InvalidBlock:
		sm.logger.WithFields(logrus.Fields{
			"height": msg.Height,
			"peer":   msg.NodeID,
		}).Warn("Handle invalid commitData")

		invalidBlockNumber.WithLabelValues("invalid_block").Inc()

		r.clearBlock()
		sm.decreaseBlockSize()

		newPeer, err := sm.pickRandomPeer(msg.NodeID)
		if err != nil {
			panic(err)
		}
		r.retryCh <- newPeer
	case common.SyncMsgType_TimeoutBlock:
		sm.logger.WithFields(logrus.Fields{
			"height": msg.Height,
			"peer":   msg.NodeID,
		}).Warn("Handle timeout block")

		invalidBlockNumber.WithLabelValues("timeout_response").Inc()

		if err := sm.addPeerTimeoutCount(msg.NodeID); err != nil {
			panic(err)
		}
		newPeer, err := sm.pickRandomPeer(msg.NodeID)
		if err != nil {
			panic(err)
		}
		r.retryCh <- newPeer
	}
}

func (sm *SyncManager) addPeerTimeoutCount(peerID string) error {
	var err error
	lo.ForEach(sm.peers, func(p *common.Peer, _ int) {
		if p.PeerID == peerID {
			p.TimeoutCount++
			if p.TimeoutCount >= sm.conf.TimeoutCountLimit {
				if empty := sm.removePeer(p.PeerID); empty {
					sm.logger.Warningf("remove peer[id:%s] err: available peer is empty, will reset peer", p.PeerID)
					sm.resetPeers()
					return
				}
			}
		}
	})
	return err
}

func (sm *SyncManager) getRequester(height uint64) *requester {
	r, loaded := sm.requesters.Load(height)
	if !loaded {
		return nil
	}
	return r.(*requester)
}

func (sm *SyncManager) pickPeer(height uint64) string {
	idx := height % uint64(len(sm.peers))
	return sm.peers[idx].PeerID
}

func (sm *SyncManager) pickRandomPeer(exceptPeerId string) (string, error) {
	if exceptPeerId != "" {
		newPeers := lo.Filter(sm.peers, func(p *common.Peer, _ int) bool {
			return p.PeerID != exceptPeerId
		})
		if len(newPeers) == 0 {
			sm.resetPeers()
			newPeers = sm.peers
		}
		return newPeers[rand.Intn(len(newPeers))].PeerID, nil
	}
	return sm.peers[rand.Intn(len(sm.peers))].PeerID, nil
}

func (sm *SyncManager) removePeer(peerId string) bool {
	var exist bool
	newPeers := lo.Filter(sm.peers, func(p *common.Peer, _ int) bool {
		if p.PeerID == peerId {
			exist = true
		}
		return p.PeerID != peerId
	})
	if !exist {
		sm.logger.WithField("peer", peerId).Warn("Remove peer failed, peer not exist")
		return false
	}

	sm.peers = newPeers
	return len(sm.peers) == 0
}

func (sm *SyncManager) resetPeers() {
	sm.peers = sm.initPeers
}

func (sm *SyncManager) Stop() {
	sm.cancel()
}

func (sm *SyncManager) stopSync() error {
	if err := sm.switchSyncStatus(false); err != nil {
		return err
	}
	sm.syncCancel()
	return nil
}

func (sm *SyncManager) Commit() chan any {
	return sm.modeConstructor.Commit()
}
