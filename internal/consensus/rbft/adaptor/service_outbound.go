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
)

func (a *RBFTAdaptor) Execute(requests []*types.Transaction, localList []bool, seqNo uint64, timestamp int64, proposerAccount string) {
	a.ReadyC <- &Ready{
		Txs:             requests,
		LocalList:       localList,
		Height:          seqNo,
		Timestamp:       timestamp,
		ProposerAccount: proposerAccount,
	}
}

func (a *RBFTAdaptor) StateUpdate(lowWatermark, seqNo uint64, digest string, checkpoints []*consensus.SignedCheckpoint, epochChanges ...*consensus.EpochChange) {
	a.StateUpdating = true
	a.StateUpdateHeight = seqNo

	var peers []string

	// get the validator set of the current local epoch
	for _, v := range a.EpochInfo.ValidatorSet {
		if v.AccountAddress != a.config.SelfAccountAddress {
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

	if chain.Height >= seqNo {
		localBlock, err := a.getBlockFunc(seqNo)
		if err != nil {
			panic("get local block failed")
		}
		if localBlock.BlockHash.String() != digest {
			a.logger.WithFields(logrus.Fields{
				"remote": digest,
				"local":  localBlock.BlockHash.String(),
				"height": seqNo,
			}).Warningf("Block hash is inconsistent in state update state, we need rollback")
			// rollback to the lowWatermark height
			startHeight = lowWatermark + 1
			latestBlockHash = localBlock.BlockHash.String()
		} else {
			a.logger.WithFields(logrus.Fields{
				"remote": digest,
				"local":  localBlock.BlockHash.String(),
				"height": seqNo,
			}).Info("state update is ignored, because we have the same block")
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

	if err := retry.Retry(func(attempt uint) error {
		err := a.sync.StartSync(peers, latestBlockHash, CalQuorum(uint64(len(peers))), startHeight, seqNo,
			checkpoints[0], epochChanges...)
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
		case <-a.quitSync:
			err := a.sync.StopSync()
			if err != nil {
				a.logger.WithFields(logrus.Fields{
					"target":      seqNo,
					"target_hash": digest,
				}).Warningf("State update stop sync failed: %v", err)
				return
			}
			a.logger.WithFields(logrus.Fields{
				"target":      seqNo,
				"target_hash": digest,
			}).Info("State update finished")
			return
		case blockCache := <-a.sync.Commit():
			a.logger.WithFields(logrus.Fields{
				"chunk start": blockCache[0].Height(),
				"chunk end":   blockCache[len(blockCache)-1].Height(),
			}).Info("fetch chunk")
			for _, block := range blockCache {
				// if the block is the target block, we should resign the stateUpdatedCheckpoint in CommitEvent
				// and send the quitSync signal to sync module
				if block.Height() == seqNo {
					stateUpdatedCheckpoint = checkpoints[0].GetCheckpoint()
					a.quitSync <- struct{}{}
				}
				a.postCommitEvent(&common.CommitEvent{
					Block:                  block,
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

func CalQuorum(N uint64) uint64 {
	f := (N - 1) / 3
	return (N + f + 2) / 2
}
