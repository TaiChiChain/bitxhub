package adaptor

import (
	"fmt"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	rbfttypes "github.com/axiomesh/axiom-bft/types"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	synccomm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/events"
)

func (a *RBFTAdaptor) Execute(requests []*types.Transaction, localList []bool, seqNo uint64, timestamp int64, proposerNodeID uint64) {
	metrics.Consensus2ExecutorBlockCounter.WithLabelValues(common.Rbft).Inc()
	metrics.BatchCommitLatency.With(prometheus.Labels{"consensus": common.Rbft, "type": "avg"}).Observe(float64(time.Now().UnixNano()-timestamp) / float64(time.Second))
	a.ReadyC <- &Ready{
		Txs:             requests,
		LocalList:       localList,
		Height:          seqNo,
		Timestamp:       timestamp,
		ProposerAccount: syscommon.StakingManagerContractAddr,
		ProposerNodeID:  proposerNodeID,
	}
}

func (a *RBFTAdaptor) StateUpdate(lowWatermark, seqNo uint64, digest string, checkpoints []*consensus.SignedCheckpoint, epochChanges ...*consensus.EpochChange) {
	a.StateUpdating = true
	a.StateUpdateHeight = seqNo

	peersM := make(map[uint64]*synccomm.Node)

	// get the validator set of the current local epoch
	for _, validatorInfo := range a.config.ChainState.ValidatorSet {
		v, err := a.config.ChainState.GetNodeInfo(validatorInfo.ID)
		if err == nil {
			if v.NodeInfo.P2PID != a.config.ChainState.SelfNodeInfo.P2PID {
				peersM[validatorInfo.ID] = &synccomm.Node{Id: validatorInfo.ID, PeerID: v.P2PID}
			}
		}
	}

	// get the validator set of the remote latest epoch
	if len(epochChanges) != 0 {
		lo.ForEach(epochChanges[len(epochChanges)-1].GetValidators().Validators, func(v *consensus.QuorumValidator, _ int) {
			if _, ok := peersM[v.Id]; !ok && v.PeerId != a.network.PeerID() {
				peersM[v.Id] = &synccomm.Node{Id: v.Id, PeerID: v.PeerId}
			}
		})
	}
	// flatten peersM
	peers := lo.Values(peersM)

	chain := a.config.ChainState.ChainMeta

	startHeight := chain.Height + 1
	latestBlockHash := chain.BlockHash.String()

	// if we had already persist last block of min epoch, dismiss the min epoch
	if len(epochChanges) != 0 {
		minEpoch := epochChanges[0]
		if minEpoch != nil && minEpoch.Checkpoint.Height() < startHeight {
			epochChanges = epochChanges[1:]
		}
	}

	if chain.Height >= seqNo {
		localBlockHeader, err := a.config.GetBlockHeaderFunc(seqNo)
		if err != nil {
			panic("get local block failed")
		}
		if localBlockHeader.Hash().String() != digest {
			a.logger.WithFields(logrus.Fields{
				"remote": digest,
				"local":  localBlockHeader.Hash().String(),
				"height": seqNo,
			}).Warningf("Block hash is inconsistent in state update state, we need rollback")
			// rollback to the lowWatermark height
			startHeight = lowWatermark + 1
			latestBlockHash = localBlockHeader.Hash().String()
		} else {
			// notify rbft report State Updated
			rbftCheckpoint := checkpoints[0].GetCheckpoint()
			a.postMockBlockEvent(&types.Block{
				Header: localBlockHeader,
			}, []*events.TxPointer{}, &common.Checkpoint{
				Epoch:  rbftCheckpoint.Epoch,
				Height: rbftCheckpoint.Height(),
				Digest: rbftCheckpoint.Digest(),
			})
			a.logger.WithFields(logrus.Fields{
				"remote": digest,
				"local":  localBlockHeader.Hash().String(),
				"height": seqNo,
			}).Info("because we have the same block," +
				" we will post mock block event to report State Updated")
			return
		}
	}

	// update the current sync height
	a.currentSyncHeight = startHeight

	a.logger.WithFields(logrus.Fields{
		"target":       a.StateUpdateHeight,
		"target_hash":  digest,
		"start":        startHeight,
		"checkpoints":  checkpoints,
		"epochChanges": epochChanges,
	}).Info("State update start")

	syncTaskDoneCh := make(chan error, 1)
	if err := retry.Retry(func(attempt uint) error {
		params := &synccomm.SyncParams{
			Peers:           peers,
			LatestBlockHash: latestBlockHash,
			// ensure sync remote count including at least one correct node
			Quorum:       common.CalFaulty(uint64(len(peers))),
			CurHeight:    startHeight,
			TargetHeight: seqNo,
			QuorumCheckpoint: &pb.QuorumCheckpoint{
				Epoch: checkpoints[0].Epoch(),
				State: &pb.ExecuteState{
					Height: checkpoints[0].Height(),
					Digest: checkpoints[0].Digest(),
				},
			},
			EpochChanges: lo.Map(epochChanges, func(epoch *consensus.EpochChange, index int) *pb.EpochChange {
				return &pb.EpochChange{
					QuorumCheckpoint: &pb.QuorumCheckpoint{
						Epoch: epoch.Checkpoint.Epoch(),
						State: &pb.ExecuteState{
							Height: epoch.Checkpoint.Height(),
							Digest: epoch.Checkpoint.Digest(),
						},
					},
				}
			}),
		}
		err := a.sync.StartSync(params, syncTaskDoneCh)
		if err != nil {
			a.logger.WithFields(logrus.Fields{
				"target":      seqNo,
				"target_hash": digest,
				"start":       startHeight,
			}).Errorf("State update start sync failed: %v", err)
			return err
		}
		return nil
	}, strategy.Limit(5), strategy.Wait(500*time.Microsecond)); err != nil {
		panic(fmt.Errorf("retry start sync failed: %v", err))
	}

	var stateUpdatedCheckpoint *common.Checkpoint
	// wait for the sync to finish
	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("state update is canceled!!!!!!")
			return
		case syncErr := <-syncTaskDoneCh:
			if syncErr != nil {
				panic(syncErr)
			}
		case <-a.quitSync:
			a.logger.WithFields(logrus.Fields{
				"target":      seqNo,
				"target_hash": digest,
			}).Info("State update finished")
			return
		case data := <-a.sync.Commit():
			blockCache, ok := data.([]synccomm.CommitData)
			if !ok {
				panic("state update failed: invalid commit data")
			}
			a.logger.WithFields(logrus.Fields{
				"chunk start": blockCache[0].GetHeight(),
				"chunk end":   blockCache[len(blockCache)-1].GetHeight(),
			}).Info("fetch chunk")
			// todo: validate epoch state not StateUpdatedCheckpoint
			for _, commitData := range blockCache {
				// if the block is the target block, we should resign the stateUpdatedCheckpoint in CommitEvent
				// and send the quitSync signal to sync module
				if commitData.GetHeight() == seqNo {
					rbftCkpt := checkpoints[0].GetCheckpoint()
					stateUpdatedCheckpoint = &common.Checkpoint{
						Epoch:  rbftCkpt.Epoch,
						Height: rbftCkpt.Height(),
						Digest: rbftCkpt.Digest(),
					}
					a.quitSync <- struct{}{}
				}
				block, ok := commitData.(*synccomm.BlockData)
				if !ok {
					panic("state update failed: invalid commit data")
				}
				a.postCommitEvent(&common.CommitEvent{
					Block:                  block.Block,
					StateUpdatedCheckpoint: stateUpdatedCheckpoint,
				})
			}
		}
	}
}

func (a *RBFTAdaptor) SendFilterEvent(_ rbfttypes.InformType, _ ...any) {
	// TODO: add implement
}

func (a *RBFTAdaptor) PostCommitEvent(commitEvent *common.CommitEvent) {
	a.postCommitEvent(commitEvent)
}

func (a *RBFTAdaptor) postCommitEvent(commitEvent *common.CommitEvent) {
	a.BlockC <- commitEvent
}

func (a *RBFTAdaptor) GetCommitChannel() chan *common.CommitEvent {
	return a.BlockC
}

func (a *RBFTAdaptor) postMockBlockEvent(block *types.Block, txHashList []*events.TxPointer, ckp *common.Checkpoint) {
	a.MockBlockFeed.Send(events.ExecutedEvent{
		Block:                  block,
		TxPointerList:          txHashList,
		StateUpdatedCheckpoint: ckp,
	})
}
