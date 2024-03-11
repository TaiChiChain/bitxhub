package adaptor

import (
	"fmt"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	rbfttypes "github.com/axiomesh/axiom-bft/types"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	sync_comm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/events"
)

func (a *RBFTAdaptor) Execute(requests []*types.Transaction, localList []bool, seqNo uint64, timestamp int64, proposerAccount string, proposerNodeID uint64) {
	a.ReadyC <- &Ready{
		Txs:             requests,
		LocalList:       localList,
		Height:          seqNo,
		Timestamp:       timestamp,
		ProposerAccount: proposerAccount,
		ProposerNodeID:  proposerNodeID,
	}
}

func (a *RBFTAdaptor) StateUpdate(lowWatermark, seqNo uint64, digest string, checkpoints []*consensus.SignedCheckpoint, epochChanges ...*consensus.EpochChange) {
	a.StateUpdating = true
	a.StateUpdateHeight = seqNo

	var peers []string

	// get the validator set of the current local epoch
	for _, v := range a.EpochInfo.ValidatorSet {
		if v.P2PNodeID != a.network.PeerID() {
			peers = append(peers, v.P2PNodeID)
		}
	}

	// get the validator set of the remote latest epoch
	if len(epochChanges) != 0 {
		peers = lo.Filter(lo.Union(peers, epochChanges[len(epochChanges)-1].GetValidators()), func(item string, idx int) bool {
			return item != a.config.Network.PeerID()
		})
	}

	chain := a.getChainMetaFunc()

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
		localBlockHeader, err := a.getBlockHeaderFunc(seqNo)
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
			a.postMockBlockEvent(&types.Block{
				Header: localBlockHeader,
			}, []*events.TxPointer{}, checkpoints[0].GetCheckpoint())
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
		"target":      a.StateUpdateHeight,
		"target_hash": digest,
		"start":       startHeight,
	}).Info("State update start")

	syncTaskDoneCh := make(chan error, 1)
	if err := retry.Retry(func(attempt uint) error {
		params := &sync_comm.SyncParams{
			Peers:            peers,
			LatestBlockHash:  latestBlockHash,
			Quorum:           CalQuorum(uint64(len(peers))),
			CurHeight:        startHeight,
			TargetHeight:     seqNo,
			QuorumCheckpoint: checkpoints[0],
			EpochChanges:     epochChanges,
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

	var stateUpdatedCheckpoint *consensus.Checkpoint
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
			blockCache, ok := data.([]sync_comm.CommitData)
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
					stateUpdatedCheckpoint = checkpoints[0].GetCheckpoint()
					a.quitSync <- struct{}{}
				}
				block, ok := commitData.(*sync_comm.BlockData)
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

func (a *RBFTAdaptor) postMockBlockEvent(block *types.Block, txHashList []*events.TxPointer, ckp *consensus.Checkpoint) {
	a.MockBlockFeed.Send(events.ExecutedEvent{
		Block:                  block,
		TxPointerList:          txHashList,
		StateUpdatedCheckpoint: ckp,
	})
}

func CalQuorum(N uint64) uint64 {
	f := (N - 1) / 3
	return (N + f + 2) / 2
}
