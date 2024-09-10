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
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	synccomm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/events"
)

func (a *RBFTAdaptor) Execute(requests []*types.Transaction, localList []bool, seqNo uint64, timestamp int64, proposerNodeID uint64) {
	a.ReadyC <- &Ready{
		Txs:             requests,
		LocalList:       localList,
		Height:          seqNo,
		Timestamp:       timestamp,
		ProposerAccount: syscommon.StakingManagerContractAddr,
		ProposerNodeID:  proposerNodeID,
	}
}

func (a *RBFTAdaptor) getPeers(stateUpdateByDiff bool, epochChanges ...*consensus.EpochChange) (map[uint64]*synccomm.Node, bool) {
	peersM := make(map[uint64]*synccomm.Node)
	var isValidator bool

	// if current node is a new archive node (data syncer), use other archive nodes as peers
	if stateUpdateByDiff {
		for _, v := range a.config.Repo.SyncArgs.RemotePeers.Validators {
			peersM[v.Id] = &synccomm.Node{Id: v.Id, PeerID: v.PeerId}
		}
	} else {
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
				if _, ok := peersM[v.Id]; !ok {
					if v.PeerId != a.network.PeerID() {
						isValidator = true
						return
					}
					peersM[v.Id] = &synccomm.Node{Id: v.Id, PeerID: v.PeerId}
				}
			})
		}
	}
	return peersM, isValidator
}

func (a *RBFTAdaptor) StateUpdate(lowWatermark, seqNo uint64, digest string, checkpoints []*consensus.SignedCheckpoint, epochChanges ...*consensus.EpochChange) {
	a.StateUpdating = true
	a.StateUpdateHeight = seqNo
	syncMode := synccomm.SyncModeFull

	chain := a.config.ChainState.ChainMeta

	startHeight := chain.Height + 1
	latestBlockHash := chain.BlockHash.String()

	// the condition for state update by diff
	stateUpdateByDiff := func() bool {
		return a.config.ChainState.IsArchiveMode && a.config.ChainState.IsDataSyncer &&
			seqNo-startHeight >= a.config.Repo.Config.Sync.ArchiveLimit
	}

	if stateUpdateByDiff() {
		syncMode = synccomm.SyncModeDiff
	}

	peersM, isValidator := a.getPeers(stateUpdateByDiff(), epochChanges...)
	// flatten peersM
	peers := lo.Values(peersM)

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
			}, []string{}, &common.Checkpoint{
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
		"mode":                synccomm.SyncModeMap[syncMode],
		"target":              a.StateUpdateHeight,
		"target_hash":         digest,
		"start":               startHeight,
		"checkpoints length":  len(checkpoints),
		"epochChanges length": len(epochChanges),
	}).Info("State update start")

	// ensure sync remote count including at least one correct node
	f := common.CalFaulty(uint64(len(peers)))
	quorum := f + 1
	if isValidator {
		// if node is a validator, it means this node is a Byzantium node,so ensure at least one correct node, quorum = f+1-1
		quorum = f + 1 - 1
	}

	syncTaskDoneCh := make(chan error, 1)
	if err := retry.Retry(func(attempt uint) error {
		params := &synccomm.SyncParams{
			Peers:           peers,
			LatestBlockHash: latestBlockHash,
			// ensure sync remote count including at least one correct node
			Quorum:           quorum,
			CurHeight:        startHeight,
			TargetHeight:     seqNo,
			QuorumCheckpoint: checkpoints[0],
			EpochChanges:     epochChanges,
		}
		if syncMode != a.sync.CurrentMode() {
			if err := a.sync.SwitchMode(syncMode); err != nil {
				panic(fmt.Errorf("switch mode from %v to %v failed: %v", a.sync.CurrentMode(), syncMode, err))
			}
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
		case <-a.quitSyncCh:
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
			for _, commitData := range blockCache {
				commitEv := &common.CommitEvent{}
				switch commitData.(type) {
				case *synccomm.BlockData:
					block, ok := commitData.(*synccomm.BlockData)
					if !ok {
						panic("state update failed: invalid commit data")
					}
					commitEv.Block = block.Block
					commitEv.SyncMeta = common.SyncMeta{
						Mode:                         synccomm.SyncModeFull,
						QuorumStateUpdatedCheckpoint: &common.Checkpoint{Height: block.GetHeight(), Digest: block.GetHash()},
					}

				case *synccomm.DiffData:
					diffData, ok := commitData.(*synccomm.DiffData)
					if !ok {
						panic("state update failed: invalid commit data")
					}
					commitEv.Block = diffData.Block
					commitEv.SyncMeta = common.SyncMeta{
						Mode:                         synccomm.SyncModeDiff,
						QuorumStateUpdatedCheckpoint: &common.Checkpoint{Height: diffData.GetHeight(), Digest: diffData.GetHash()},
						StateJournal:                 diffData.StateJournal,
						Receipts:                     diffData.Receipts,
					}
				}
				a.postCommitEvent(commitEv)
			}
			if blockCache[len(blockCache)-1].GetHeight() == seqNo {
				a.quitSyncCh <- struct{}{}
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

func (a *RBFTAdaptor) postMockBlockEvent(block *types.Block, txHashList []string, ckp *common.Checkpoint) {
	a.MockBlockFeed.Send(events.ExecutedEvent{
		Block:      block,
		TxHashList: txHashList,
	})
}
