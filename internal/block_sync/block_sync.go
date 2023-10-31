package block_sync

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
	"github.com/gammazero/workerpool"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

//go:generate mockgen -destination mock_block_sync/mock_block_sync.go -package mock_block_sync -source block_sync.go
type Sync interface {
	Start() error
	Stop()
	Commit() chan []*types.Block
	StartSync(peers []string, curBlockHash string, quorum, curHeight, targetHeight uint64, quorumCheckpoint *consensus.SignedCheckpoint, epc ...*consensus.EpochChange) error
	StopSync() error
}

// todo: requester and chunk change to sync pool
type BlockSync struct {
	conf               repo.Sync
	syncStatus         atomic.Bool         // sync status
	peers              []*peer             // p2p set of latest epoch validatorSet
	quorum             uint64              // quorum of latest epoch validatorSet
	curHeight          uint64              // current block which we need sync
	recvBlockSize      atomic.Int64        // current chunk had received block size
	latestCheckedState *pb.CheckpointState // latest checked block state
	targetHeight       uint64              // sync target block height
	requesters         sync.Map            // requester map
	requesterLen       atomic.Int64        // requester length

	quorumCheckpoint      *consensus.SignedCheckpoint // latest checkpoint from remote
	epochChanges          []*consensus.EpochChange    // every epoch change which the node behind
	getBlockFunc          func(height uint64) (*types.Block, error)
	network               network.Network
	syncBlockRequestPipe  network.Pipe
	syncBlockResponsePipe network.Pipe

	blockCache   []*types.Block      // restore block temporary
	blockCacheCh chan []*types.Block // restore block temporary

	chunk            *chunk                 // every chunk task
	recvStateCh      chan *wrapperStateResp // receive state from remote peer
	quitStateCh      chan bool              // quit state channel
	stateTaskDone    atomic.Bool            // state task done signal
	validChunkTaskCh chan struct{}          // start validate chunk task singal
	syncTaskDone     atomic.Bool            // all chunk task done signal

	invalidRequestCh chan *invalidMsg // timeout or invalid of sync Block request

	ctx    context.Context
	cancel context.CancelFunc

	syncCtx    context.Context
	syncCancel context.CancelFunc

	logger logrus.FieldLogger
}

func NewBlockSync(logger logrus.FieldLogger, fn func(height uint64) (*types.Block, error), network network.Network, cnf repo.Sync) (*BlockSync, error) {
	ctx, cancel := context.WithCancel(context.Background())
	blockSync := &BlockSync{
		logger:           logger,
		invalidRequestCh: make(chan *invalidMsg, 1024),
		recvStateCh:      make(chan *wrapperStateResp, 1024),
		blockCacheCh:     make(chan []*types.Block, 1024),
		validChunkTaskCh: make(chan struct{}, 1),
		quitStateCh:      make(chan bool, 1),
		getBlockFunc:     fn,
		network:          network,
		conf:             cnf,

		ctx:    ctx,
		cancel: cancel,
	}

	// init syncStatus
	blockSync.syncStatus.Store(false)

	// init sync block pipe
	reqPipe, err := blockSync.network.CreatePipe(blockSync.ctx, syncBlockRequestPipe)
	if err != nil {
		return nil, err
	}
	blockSync.syncBlockRequestPipe = reqPipe

	respPipe, err := blockSync.network.CreatePipe(blockSync.ctx, syncBlockResponsePipe)
	if err != nil {
		return nil, err
	}
	blockSync.syncBlockResponsePipe = respPipe

	blockSync.logger.Info("Init block sync success")

	return blockSync, nil
}

func (bs *BlockSync) StartSync(peers []string, latestBlockHash string, quorum, curHeight, targetHeight uint64,
	quorumCheckpoint *consensus.SignedCheckpoint, epc ...*consensus.EpochChange) error {
	now := time.Now()
	syncCount := targetHeight - curHeight + 1
	syncCtx, syncCancel := context.WithCancel(context.Background())
	bs.syncCtx = syncCtx
	bs.syncCancel = syncCancel

	// 1. update block sync info
	bs.InitBlockSyncInfo(peers, latestBlockHash, quorum, curHeight, targetHeight, quorumCheckpoint, epc...)

	// 2. send sync state request to all validators, waiting for quorum response
	err := bs.requestSyncState(bs.curHeight-1, latestBlockHash)
	if err != nil {
		return err
	}

	bs.logger.WithFields(logrus.Fields{
		"quorum": bs.quorum,
	}).Info("Receive quorum response")

	// 3. switch sync status to true, if switch failed, return error
	if err := bs.switchSyncStatus(true); err != nil {
		return err
	}

	// 4. start listen sync block response
	go bs.listenSyncBlockResponse()

	// 5. start sync block task
	go func() {
		for {
			select {
			case <-bs.syncCtx.Done():
				return
			case msg := <-bs.invalidRequestCh:
				bs.handleInvalidRequest(msg)

			case <-bs.validChunkTaskCh:
				bs.logger.WithFields(logrus.Fields{
					"start":  bs.curHeight,
					"target": bs.curHeight + bs.chunk.chunkSize - 1,
				}).Info("chunk task has done")

				invalidReqs, err := bs.validateChunk()
				if err != nil {
					bs.logger.WithFields(logrus.Fields{
						"err": err,
					}).Error("Validate chunk failed")
					panic(err)
				}

				if len(invalidReqs) != 0 {
					lo.ForEach(invalidReqs, func(req *invalidMsg, index int) {
						bs.invalidRequestCh <- req
						bs.logger.WithFields(logrus.Fields{
							"peer":   req.nodeID,
							"height": req.height,
							"err":    req.errMsg,
						}).Warning("Receive Invalid block")
					})
					continue
				}

				lastR := bs.getRequester(bs.curHeight + bs.chunk.chunkSize - 1)
				if lastR == nil {
					bs.logger.WithFields(logrus.Fields{
						"height": bs.curHeight + bs.chunk.chunkSize - 1,
					}).Error("Load requester failed")
					continue
				}
				err = bs.validateChunkState(lastR.block.Height(), lastR.block.BlockHash.String())
				if err != nil {
					bs.logger.WithFields(logrus.Fields{
						"err": err,
					}).Error("Validate last chunk state failed")
					continue
				}

				// release requester and send block to blockCacheCh
				bs.requesters.Range(func(height, r any) bool {
					if r.(*requester).block == nil {
						bs.invalidRequestCh <- &invalidMsg{
							nodeID: r.(*requester).peerID,
							height: height.(uint64),
							typ:    syncMsgType_TimeoutBlock,
						}
						return false
					}

					bs.blockCache[height.(uint64)-bs.curHeight] = r.(*requester).block
					return true
				})

				// if blockCache is not full, continue to receive block
				if len(bs.blockCache) != int(bs.chunk.chunkSize) {
					continue
				}

				bs.updateLatestCheckedState(bs.blockCache[len(bs.blockCache)-1].Height(), bs.blockCache[len(bs.blockCache)-1].BlockHash.String())

				// if valid chunk task done, release all requester
				lo.ForEach(bs.blockCache, func(block *types.Block, index int) {
					bs.releaseRequester(block.Height())
				})

				if bs.chunk.checkPoint != nil {
					idx := int(bs.chunk.checkPoint.Height - bs.curHeight)
					if idx < 0 || idx > len(bs.blockCache)-1 {
						bs.logger.Errorf("chunk checkpoint index out of range, checkpoint height:%d, current Height:%d, "+
							"blockCache len:%d", bs.chunk.checkPoint.Height, bs.curHeight, len(bs.blockCache))
						continue
					}

					// if checkpoint is not equal to last block, it means we sync wrong block, panic it
					if err = bs.verifyChunkCheckpoint(bs.blockCache[idx]); err != nil {
						bs.logger.Errorf("verify chunk checkpoint failed: %s", err)
						panic(err)
					}
				}

				bs.blockCacheCh <- bs.blockCache

				if bs.curHeight+bs.chunk.chunkSize-1 == bs.targetHeight {
					bs.syncTaskDone.Store(true)
					bs.logger.WithFields(logrus.Fields{
						"count":  syncCount,
						"target": bs.targetHeight,
						"elapse": time.Since(now).Seconds(),
					}).Info("Block sync done")
					blockSyncDuration.WithLabelValues(strconv.Itoa(int(syncCount))).Observe(time.Since(now).Seconds())
				} else {
					// update chunkSize and curHeight
					bs.updateStatus()
				}
			default:
				switch {
				case bs.syncTaskDone.Load():
					// if sync task done, waiting for quit signal
					continue
				case bs.requesterLen.Load() >= int64(bs.chunk.chunkSize):
					// sleep a while to wait for chunk task done, but not block sync task
					time.Sleep(requestInterval)
					continue
				default:
					bs.makeRequesters(bs.curHeight + uint64(bs.requesterLen.Load()))
				}
			}
		}
	}()

	return nil
}

func (bs *BlockSync) validateChunkState(localHeight uint64, localHash string) error {
	err := bs.requestSyncState(localHeight, localHash)
	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockSync) validateChunk() ([]*invalidMsg, error) {
	parentHash := bs.latestCheckedState.Digest
	for i := bs.curHeight; i < bs.curHeight+bs.chunk.chunkSize; i++ {
		r := bs.getRequester(i)
		if r == nil {
			return nil, fmt.Errorf("requester[height:%d] is nil", i)
		}
		block := r.block
		if block == nil {
			bs.logger.WithFields(logrus.Fields{
				"height": i,
			}).Error("Block is nil")
			return []*invalidMsg{
				{
					nodeID: r.peerID,
					height: i,
					typ:    syncMsgType_TimeoutBlock,
				},
			}, nil
		}

		if block.BlockHeader.ParentHash.String() != parentHash {
			bs.logger.WithFields(logrus.Fields{
				"height":               i,
				"expect parent hash":   parentHash,
				"expect parent height": i - 1,
				"actual parent hash":   block.BlockHeader.ParentHash.String(),
			}).Error("Block parent hash is not equal to latest checked state")

			invalidMsgs := make([]*invalidMsg, 0)

			// if we have not previous requester, it means we had already checked previous block,
			// so we just return current invalid block
			prevR := bs.getRequester(i - 1)
			if prevR != nil {
				// we are not sure which block is wrong,
				// maybe parent block has wrong hash, maybe current block has wrong parent hash, so we return two invalidMsg
				prevInvalidMsg := &invalidMsg{
					nodeID: prevR.peerID,
					height: i - 1,
					typ:    syncMsgType_InvalidBlock,
				}
				invalidMsgs = append(invalidMsgs, prevInvalidMsg)
			}
			invalidMsgs = append(invalidMsgs, &invalidMsg{
				nodeID: r.peerID,
				height: i,
				typ:    syncMsgType_InvalidBlock,
			})

			return invalidMsgs, nil
		}

		parentHash = block.BlockHash.String()
	}
	return nil, nil
}

func (bs *BlockSync) updateLatestCheckedState(height uint64, digest string) {
	bs.latestCheckedState = &pb.CheckpointState{
		Height: height,
		Digest: digest,
	}
}

func (bs *BlockSync) InitBlockSyncInfo(peers []string, latestBlockHash string, quorum, curHeight, targetHeight uint64,
	quorumCheckpoint *consensus.SignedCheckpoint, epc ...*consensus.EpochChange) {
	bs.peers = make([]*peer, len(peers))
	lo.ForEach(peers, func(p string, index int) {
		bs.peers[index] = &peer{
			peerID:       p,
			timeoutCount: 0,
		}
	})
	bs.quorum = quorum
	bs.curHeight = curHeight
	bs.targetHeight = targetHeight
	bs.quorumCheckpoint = quorumCheckpoint
	bs.epochChanges = epc
	bs.recvBlockSize.Store(0)
	bs.syncTaskDone.Store(false)

	bs.updateLatestCheckedState(curHeight-1, latestBlockHash)

	// init chunk
	bs.initChunk()

	bs.blockCache = make([]*types.Block, bs.chunk.chunkSize)
}

func (bs *BlockSync) initChunk() {
	chunkSize := bs.targetHeight - bs.curHeight + 1
	if chunkSize > bs.conf.ConcurrencyLimit {
		chunkSize = bs.conf.ConcurrencyLimit
	}

	// if we have epoch change, chunk size need smaller than epoch size
	if len(bs.epochChanges) != 0 {
		epochSize := bs.epochChanges[0].GetCheckpoint().Checkpoint.Height() - bs.curHeight + 1
		if epochSize < chunkSize {
			chunkSize = epochSize
		}
	}

	bs.chunk = &chunk{
		chunkSize: chunkSize,
	}

	chunkMaxHeight := bs.curHeight + chunkSize - 1

	var chunkCheckpoint *pb.CheckpointState

	if len(bs.epochChanges) != 0 {
		chunkCheckpoint = &pb.CheckpointState{
			Height: bs.epochChanges[0].GetCheckpoint().Checkpoint.Height(),
			Digest: bs.epochChanges[0].GetCheckpoint().Checkpoint.Digest(),
		}
	} else {
		chunkCheckpoint = &pb.CheckpointState{
			Height: bs.quorumCheckpoint.Height(),
			Digest: bs.quorumCheckpoint.Digest(),
		}
	}
	bs.chunk.fillCheckPoint(chunkMaxHeight, chunkCheckpoint)
}

func (bs *BlockSync) switchSyncStatus(status bool) error {
	if bs.syncStatus.Load() == status {
		return fmt.Errorf("status is already %v", status)
	}
	bs.syncStatus.Store(status)
	bs.logger.Info("SwitchSyncStatus: status is ", status)
	return nil
}

func (bs *BlockSync) listenSyncStateResp(ctx context.Context, cancel context.CancelFunc, height uint64, localHash string) {
	diffState := make(map[string][]string)

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-bs.recvStateCh:
			bs.handleSyncStateResp(resp, diffState, height, localHash, cancel)
		}
	}
}

func (bs *BlockSync) requestSyncState(height uint64, localHash string) error {
	bs.logger.WithFields(logrus.Fields{
		"height": height,
	}).Info("Start request sync state")
	bs.stateTaskDone.Store(false)

	// 1. start listen sync state response
	stateCtx, stateCancel := context.WithCancel(context.Background())
	go bs.listenSyncStateResp(stateCtx, stateCancel, height, localHash)

	wp := workerpool.New(len(bs.peers))
	// send sync state request to all validators, check our local state(latest block) is equal to quorum state
	req := &pb.SyncStateRequest{
		Height: height,
	}
	data, err := req.MarshalVT()
	if err != nil {
		return err
	}

	// 2. send sync state request to all validators asynchronously
	lo.ForEach(bs.peers, func(p *peer, index int) {
		select {
		case <-stateCtx.Done():
			wp.Stop()
			bs.logger.Debug("receive quit signal, Quit request state")
			return
		default:
			wp.Submit(func() {
				if err = retry.Retry(func(attempt uint) error {
					select {
					case <-stateCtx.Done():
						bs.logger.WithFields(logrus.Fields{
							"peer":   p.peerID,
							"height": height,
						}).Debug("receive quit signal, Quit request state")
						return nil
					default:
						bs.logger.WithFields(logrus.Fields{
							"peer":   p.peerID,
							"height": height,
						}).Debug("start send sync state request")
						resp, err := bs.network.Send(p.peerID, &pb.Message{
							Type: pb.Message_SYNC_STATE_REQUEST,
							Data: data,
						})
						if err != nil {
							bs.logger.WithFields(logrus.Fields{
								"peer": p.peerID,
								"err":  err,
							}).Warn("Send sync state request failed")
							return err
						}

						if err = bs.isValidSyncResponse(resp, p.peerID); err != nil {
							bs.logger.WithFields(logrus.Fields{
								"peer": p.peerID,
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
							bs.logger.WithFields(logrus.Fields{
								"peer":   p.peerID,
								"height": height,
							}).Debug("receive quit signal, Quit request state")
							return nil
						default:
							hash := sha256.Sum256(resp.Data)
							bs.recvStateCh <- &wrapperStateResp{
								peerID: resp.From,
								hash:   types.NewHash(hash[:]).String(),
								resp:   stateResp,
							}
						}
						return nil
					}
				}, strategy.Limit(10), strategy.Backoff(backoff.Fibonacci(500*time.Millisecond))); err != nil {
					bs.logger.Errorf("Retry send sync state request failed: %s", err)
					return
				}
				bs.logger.WithFields(logrus.Fields{
					"peer":   p.peerID,
					"height": height,
				}).Debug("Send sync state request success")
			})
		}
	})

	// 3. waiting for quorum response
	ticker := time.NewTicker(bs.conf.WaitStateTimeout.ToDuration())
	defer ticker.Stop()
	for {
		select {
		case success := <-bs.quitStateCh:
			if !success {
				return fmt.Errorf("receive invalid state response: height:%d", height)
			}
			return nil

		case <-ticker.C:
			return fmt.Errorf("timeout send sync request: height:%d", height)
		}
	}
}

func (bs *BlockSync) isValidSyncResponse(msg *pb.Message, id string) error {
	if msg == nil || msg.Data == nil {
		return errors.New("sync response is nil")
	}

	if msg.From != id {
		return fmt.Errorf("receive different peer sync response, expect peer id is %s,"+
			" but receive peer id is %s", id, msg.From)
	}

	return nil
}

func (bs *BlockSync) listenSyncBlockRequest() {
	for {
		msg := bs.syncBlockRequestPipe.Receive(bs.ctx)
		if msg == nil {
			bs.logger.Info("Stop listen sync block request")
			return
		}
		req := &pb.SyncBlockRequest{}
		if err := req.UnmarshalVT(msg.Data); err != nil {
			bs.logger.Errorf("Unmarshal sync block request failed: %s", err)
			continue
		}
		block, err := bs.getBlockFunc(req.Height)
		if err != nil {
			bs.logger.WithFields(logrus.Fields{
				"from": msg.From,
				"err":  err,
			}).Error("Get block failed")
			continue
		}

		blockBytes, err := block.Marshal()
		if err != nil {
			bs.logger.Errorf("Marshal block failed: %s", err)
			continue
		}
		resp := &pb.Message{
			Type: pb.Message_SYNC_BLOCK_RESPONSE,
			Data: blockBytes,
		}
		data, err := resp.MarshalVT()
		if err != nil {
			bs.logger.Errorf("Marshal sync block response failed: %s", err)
			continue
		}
		if err = retry.Retry(func(attempt uint) error {
			err = bs.syncBlockResponsePipe.Send(bs.ctx, msg.From, data)
			if err != nil {
				bs.logger.WithFields(logrus.Fields{
					"from": msg.From,
					"err":  err,
				}).Error("Send sync block response failed")
				return err
			}
			bs.logger.WithFields(logrus.Fields{
				"from":   msg.From,
				"height": block.Height(),
			}).Debug("Send sync block response success")
			return nil
		}, strategy.Limit(maxRetryCount), strategy.Wait(500*time.Millisecond)); err != nil {
			bs.logger.Errorf("Retry send sync block response failed: %s", err)

			continue
		}
	}
}

func (bs *BlockSync) listenSyncBlockResponse() {
	for {
		select {
		case <-bs.syncCtx.Done():
			return
		default:
			msg := bs.syncBlockResponsePipe.Receive(bs.syncCtx)
			if msg == nil {
				return
			}

			p2pMsg := &pb.Message{}
			if err := p2pMsg.UnmarshalVT(msg.Data); err != nil {
				bs.logger.Errorf("Unmarshal sync block response failed: %s", err)
				continue
			}
			if p2pMsg.Type != pb.Message_SYNC_BLOCK_RESPONSE {
				bs.logger.Errorf("Receive invalid sync block response type: %s", p2pMsg.Type)
				continue
			}
			block := &types.Block{}
			if err := block.Unmarshal(p2pMsg.Data); err != nil {
				bs.logger.Errorf("Unmarshal block failed: %s", err)
				continue
			}

			if err := bs.addBlock(block, msg.From); err != nil {
				bs.logger.WithFields(logrus.Fields{
					"from": msg.From,
					"err":  err,
				}).Error("Add block failed")
				continue
			}

			if bs.collectChunkTaskDone() {
				bs.logger.WithFields(logrus.Fields{
					"latest block": block.Height(),
					"hash":         block.Hash(),
					"peer":         msg.From,
				}).Debug("Receive chunk block success")
				// send valid chunk task signal
				bs.validChunkTaskCh <- struct{}{}
			}
		}
	}
}

func (bs *BlockSync) verifyChunkCheckpoint(checkBlock *types.Block) error {
	if bs.chunk.checkPoint.Digest != checkBlock.BlockHash.String() {
		return fmt.Errorf("quorum checkpoint[height:%d hash:%s] is not equal to current hash[%s]",
			bs.chunk.checkPoint.Height, bs.chunk.checkPoint.Digest, checkBlock.BlockHash.String())
	}
	return nil
}

func (bs *BlockSync) addBlock(block *types.Block, from string) error {
	req := bs.getRequester(block.Height())
	if req == nil {
		return fmt.Errorf("requester[height:%d] is nil", block.Height())
	}

	if req.peerID != from {
		return fmt.Errorf("receive block which not distribute requester, height:%d, "+
			"receive from:%s, expect from:%s", block.Height(), from, req.peerID)
	}

	if req.block == nil {
		req.setBlock(block)
		bs.increaseBlockSize()
	}
	bs.logger.WithFields(logrus.Fields{
		"height":    block.Height(),
		"from":      from,
		"add_block": block.BlockHash.String(),
		"hash":      req.block.BlockHash.String(),
	}).Debug("Receive block success")
	return nil
}

func (bs *BlockSync) collectChunkTaskDone() bool {
	if bs.chunk.chunkSize == 0 {
		return true
	}

	return bs.recvBlockSize.Load() >= int64(bs.chunk.chunkSize)
}

func (bs *BlockSync) handleSyncStateResp(msg *wrapperStateResp, diffState map[string][]string, localHeight uint64,
	localHash string, cancel context.CancelFunc) {
	bs.logger.WithFields(logrus.Fields{
		"peer":   msg.peerID,
		"height": msg.resp.CheckpointState.Height,
		"digest": msg.resp.CheckpointState.Digest,
	}).Debug("Receive sync state response")

	if bs.stateTaskDone.Load() || localHeight != msg.resp.CheckpointState.Height {
		bs.logger.WithFields(logrus.Fields{
			"peer":   msg.peerID,
			"height": msg.resp.CheckpointState.Height,
			"digest": msg.resp.CheckpointState.Digest,
		}).Debug("Receive state response after state task done, we ignore it")
		return
	}
	diffState[msg.hash] = append(diffState[msg.hash], msg.peerID)

	// if quorum state is enough, update quorum state
	if len(diffState[msg.hash]) >= int(bs.quorum) {
		defer cancel()
		if msg.resp.CheckpointState.Digest != localHash {
			// if we have not started sync, we need panic
			if !bs.syncStatus.Load() {
				panic(fmt.Errorf("local block [height:%d,hash:%s] is not equal to quorum state[height:%d, hash:%s]",
					localHeight, localHash, msg.resp.CheckpointState.Height, msg.resp.CheckpointState.Digest))
			} else {
				r := bs.getRequester(msg.resp.CheckpointState.Height)
				if r == nil {
					panic(fmt.Errorf("get state of requester[height:%d] is nil", msg.resp.CheckpointState.Height))
				}
				bs.logger.WithFields(logrus.Fields{
					"peer":              r.peerID,
					"height":            msg.resp.CheckpointState.Height,
					"local":             localHeight,
					"localHeight":       localHash,
					"remote checkpoint": msg.resp.CheckpointState.Digest,
				}).Warn("Receive invalid state response, retry request block")
				bs.invalidRequestCh <- &invalidMsg{
					nodeID: r.peerID,
					height: msg.resp.CheckpointState.Height,
					typ:    syncMsgType_InvalidBlock,
				}
				bs.quitState(false)
				return
			}
		}

		delete(diffState, msg.hash)
		// remove peers which not in quorum state
		if len(diffState) != 0 {
			wrongPeers := lo.Values(diffState)
			lo.ForEach(lo.Flatten(wrongPeers), func(peer string, _ int) {
				if empty := bs.removePeer(peer); empty {
					panic("available peer is empty")
				}
			})
		}

		// todo: In cases of network latency, there is an small probability that not reaching a quorum of same state.
		// For example, if the validator set is 4, and the quorum is 3:
		// 1. if the current node is forked,
		// 2. validator need send state which obtaining low watermark height block,
		// 3. validator have different low watermark height block due to network latency,
		// 4. it can lead to state inconsistency, and the node will be stuck in the state sync process.
		bs.logger.Debug("Receive quorum state from peers")
		bs.quitState(true)
	}
}

func (bs *BlockSync) Start() error {
	// register message handler
	err := bs.network.RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, bs.handleSyncState)
	if err != nil {
		return err
	}
	// start handle sync block request
	go bs.listenSyncBlockRequest()
	bs.logger.Info("Start listen sync request")

	return nil
}

func (bs *BlockSync) makeRequesters(height uint64) {
	peerID := bs.pickPeer(height)
	request := newRequester(bs.ctx, peerID, height, bs.invalidRequestCh, bs.syncBlockRequestPipe)
	bs.increaseRequester(request, height)
	request.start(bs.conf.RequesterRetryTimeout.ToDuration())
}

// todo: add metrics
func (bs *BlockSync) increaseRequester(r *requester, height uint64) {
	oldR, loaded := bs.requesters.LoadOrStore(height, r)
	if !loaded {
		bs.requesterLen.Add(1)
		requesterNumber.Inc()
	} else {
		bs.logger.WithFields(logrus.Fields{
			"height": height,
		}).Warn("Make requester Error, requester is not nil, we will reset the old requester")
		oldR.(*requester).quitCh <- struct{}{}
	}
}

func (bs *BlockSync) updateStatus() {
	bs.curHeight += bs.chunk.chunkSize
	if len(bs.epochChanges) != 0 {
		bs.epochChanges = bs.epochChanges[1:]
	}
	bs.initChunk()
	bs.resetBlockSize()

	bs.blockCache = make([]*types.Block, bs.chunk.chunkSize)
}

// todo: add metrics
func (bs *BlockSync) increaseBlockSize() {
	bs.recvBlockSize.Add(1)
	recvBlockNumber.WithLabelValues(strconv.Itoa(int(bs.chunk.chunkSize))).Inc()
}

func (bs *BlockSync) decreaseBlockSize() {
	bs.recvBlockSize.Add(-1)
	recvBlockNumber.WithLabelValues(strconv.Itoa(int(bs.chunk.chunkSize))).Dec()
}

func (bs *BlockSync) resetBlockSize() {
	bs.recvBlockSize.Store(0)
	bs.logger.WithFields(logrus.Fields{
		"blockSize": bs.recvBlockSize.Load(),
	}).Debug("Reset block size")
	recvBlockNumber.WithLabelValues(strconv.Itoa(int(bs.chunk.chunkSize))).Set(0)
}

func (bs *BlockSync) quitState(success bool) {
	bs.quitStateCh <- success
	bs.stateTaskDone.Store(true)
}

func (bs *BlockSync) releaseRequester(height uint64) {
	r, loaded := bs.requesters.LoadAndDelete(height)
	if !loaded {
		bs.logger.WithFields(logrus.Fields{
			"height": height,
		}).Warn("Release requester Error, requester is nil")
	} else {
		bs.requesterLen.Add(-1)
		requesterNumber.Dec()
	}
	r.(*requester).quitCh <- struct{}{}
}

func (bs *BlockSync) handleInvalidRequest(msg *invalidMsg) {
	// retry request
	r := bs.getRequester(msg.height)
	if r == nil {
		bs.logger.Errorf("Retry request block Error, requester[height:%d] is nil", msg.height)
		return
	}
	switch msg.typ {
	case syncMsgType_ErrorMsg:
		bs.logger.WithFields(logrus.Fields{
			"height": msg.height,
			"peer":   msg.nodeID,
			"err":    msg.errMsg,
		}).Warn("Handle error msg Block")

		invalidBlockNumber.WithLabelValues("send_request_err").Inc()
		newPeer, err := bs.pickRandomPeer(msg.nodeID)
		if err != nil {
			panic(err)
		}

		r.retryCh <- newPeer

	case syncMsgType_InvalidBlock:
		bs.logger.WithFields(logrus.Fields{
			"height": msg.height,
			"peer":   msg.nodeID,
		}).Warn("Handle invalid block")

		invalidBlockNumber.WithLabelValues("invalid_block").Inc()

		r.clearBlock()
		bs.decreaseBlockSize()

		newPeer, err := bs.pickRandomPeer(msg.nodeID)
		if err != nil {
			panic(err)
		}
		r.retryCh <- newPeer
	case syncMsgType_TimeoutBlock:
		bs.logger.WithFields(logrus.Fields{
			"height": msg.height,
			"peer":   msg.nodeID,
		}).Warn("Handle timeout block")

		invalidBlockNumber.WithLabelValues("timeout_response").Inc()

		if err := bs.addPeerTimeoutCount(msg.nodeID); err != nil {
			panic(err)
		}
		newPeer, err := bs.pickRandomPeer(msg.nodeID)
		if err != nil {
			panic(err)
		}
		r.retryCh <- newPeer
	}
}

func (bs *BlockSync) addPeerTimeoutCount(peerID string) error {
	var err error
	lo.ForEach(bs.peers, func(p *peer, _ int) {
		if p.peerID == peerID {
			p.timeoutCount++
			if p.timeoutCount >= bs.conf.TimeoutCountLimit {
				if empty := bs.removePeer(p.peerID); empty {
					err = fmt.Errorf("remove peer[id:%s] err: available peer is empty", p.peerID)
					bs.logger.Errorf(err.Error())
					return
				}
			}
		}
	})
	return err
}

func (bs *BlockSync) getRequester(height uint64) *requester {
	r, loaded := bs.requesters.Load(height)
	if !loaded {
		return nil
	}
	return r.(*requester)
}

func (bs *BlockSync) pickPeer(height uint64) string {
	idx := height % uint64(len(bs.peers))
	return bs.peers[idx].peerID
}

func (bs *BlockSync) pickRandomPeer(exceptPeerId string) (string, error) {
	if exceptPeerId != "" {
		newPeers := lo.Filter(bs.peers, func(p *peer, _ int) bool {
			return p.peerID != exceptPeerId
		})
		if len(newPeers) == 0 {
			return "", fmt.Errorf("no peer except %s", exceptPeerId)
		}
		return newPeers[rand.Intn(len(newPeers))].peerID, nil
	}
	return bs.peers[rand.Intn(len(bs.peers))].peerID, nil
}

func (bs *BlockSync) removePeer(peerId string) bool {
	var exist bool
	newPeers := lo.Filter(bs.peers, func(p *peer, _ int) bool {
		if p.peerID == peerId {
			exist = true
		}
		return p.peerID != peerId
	})
	if !exist {
		bs.logger.WithField("peer", peerId).Warn("Remove peer failed, peer not exist")
		return false
	}

	bs.peers = newPeers
	return len(bs.peers) == 0
}

func (bs *BlockSync) Stop() {
	bs.cancel()
}

func (bs *BlockSync) StopSync() error {
	if err := bs.switchSyncStatus(false); err != nil {
		return err
	}
	bs.syncCancel()
	return nil
}

func (bs *BlockSync) Commit() chan []*types.Block {
	return bs.blockCacheCh
}
