package adaptor

import (
	"fmt"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	rbft "github.com/axiomesh/axiom-bft/common/consensus"
	rbfttypes "github.com/axiomesh/axiom-bft/types"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	synccomm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type (
	ValidatorID = uint64
	Signature   = []byte
)

func (a *RBFTAdaptor) Execute(requests []*types.Transaction, localList []bool, seqNo uint64, timestamp int64, proposerNodeID uint64) {
	metrics.Consensus2ExecutorBlockCounter.WithLabelValues(consensustypes.Rbft).Inc()
	metrics.BatchCommitLatency.With(prometheus.Labels{"consensus": consensustypes.Rbft, "type": "avg"}).Observe(float64(time.Now().UnixNano()-timestamp) / float64(time.Second))
	a.ReadyC <- &Ready{
		Txs:             requests,
		LocalList:       localList,
		Height:          seqNo,
		Timestamp:       timestamp,
		ProposerAccount: syscommon.StakingManagerContractAddr,
		ProposerNodeID:  proposerNodeID,
	}
}

func (a *RBFTAdaptor) StateUpdate(lowWatermark, seqNo uint64, digest string, checkpoints []*rbft.SignedCheckpoint, epochChanges ...*rbft.EpochChange) {
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

	var isValidator bool
	// get the validator set of the remote latest epoch
	if len(epochChanges) != 0 {
		lo.ForEach(epochChanges[len(epochChanges)-1].GetValidators().Validators, func(v *rbft.QuorumValidator, _ int) {
			if _, ok := peersM[v.Id]; !ok {
				if v.PeerId != a.network.PeerID() {
					isValidator = true
					return
				}
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
			}, []*events.TxPointer{}, &consensustypes.Checkpoint{
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
			Quorum:          quorum,
			CurHeight:       startHeight,
			TargetHeight:    seqNo,
			QuorumCheckpoint: &rbft.RbftQuorumCheckpoint{
				QuorumCheckpoint: &rbft.QuorumCheckpoint{
					Checkpoint: checkpoints[0].GetCheckpoint(),
				},
			},
			EpochChanges: lo.Map(epochChanges, func(epochChange *rbft.EpochChange, index int) *consensustypes.EpochChange {
				return &consensustypes.EpochChange{
					QuorumCheckpoint: &rbft.RbftQuorumCheckpoint{QuorumCheckpoint: epochChange.GetCheckpoint()},
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

	var stateUpdatedCheckpoint *consensustypes.Checkpoint
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
					stateUpdatedCheckpoint = &consensustypes.Checkpoint{
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
				a.postCommitEvent(&consensustypes.CommitEvent{
					Block:                  block.Block,
					StateUpdatedCheckpoint: stateUpdatedCheckpoint,
				})
			}
		}
	}
}

// SendFilterEvent posts some impotent events to application layer.
// Users can decide to post filer event synchronously or asynchronously.
func (a *RBFTAdaptor) SendFilterEvent(informType rbfttypes.InformType, message ...any) {
	if informType == rbfttypes.InformTypeFilterStableCheckpoint {
		if len(message) != 1 {
			a.logger.Errorf("length: %d", len(message))
			return
		}
		signedCheckpoints, ok := message[0].([]*rbft.SignedCheckpoint)
		if !ok {
			a.logger.Error("invalid format")
			return
		}
		a.Checkpoint(signedCheckpoints)
	}
}

func (a *RBFTAdaptor) PostCommitEvent(commitEvent *consensustypes.CommitEvent) {
	a.postCommitEvent(commitEvent)
}

func (a *RBFTAdaptor) postCommitEvent(commitEvent *consensustypes.CommitEvent) {
	a.BlockC <- commitEvent
}

func (a *RBFTAdaptor) GetCommitChannel() chan *consensustypes.CommitEvent {
	return a.BlockC
}

func (a *RBFTAdaptor) postMockBlockEvent(block *types.Block, txHashList []*events.TxPointer, ckp *consensustypes.Checkpoint) {
	a.MockBlockFeed.Send(events.ExecutedEvent{
		Block:                  block,
		TxPointerList:          txHashList,
		StateUpdatedCheckpoint: ckp,
	})
}

func (a *RBFTAdaptor) Checkpoint(checkpoints []*rbft.SignedCheckpoint) {
	if len(checkpoints) == 0 {
		a.logger.Error("invalid signed checkpoint")
		return
	}

	height := checkpoints[0].Checkpoint.Height()
	if height <= a.stableHeight {
		a.logger.Debugf("need not handle quorum checkpoint which height %d is smaller than stable height: %d", height, a.stableHeight)
		return

	}

	var checkpoint *rbft.Checkpoint
	uniqueSigs := make(map[ValidatorID]Signature, len(checkpoints))
	for _, signed := range checkpoints {
		if checkpoint == nil {
			checkpoint = signed.Checkpoint
		} else {
			// todo(lrx): slash other nodes if checkpoint is not consistent?
			if !checkpoint.Equals(signed.Checkpoint) {
				a.logger.Errorf("inconsistent checkpoint, one: %+v, another: %+v",
					checkpoint, signed.Checkpoint)
				return
			}
		}
		_, exist := uniqueSigs[signed.GetAuthor()]
		// todo(lrx): slash other nodes if duplicate signature?
		if exist {
			a.logger.Errorf("duplicate signature in checkpoint: %d", signed.GetAuthor())
			return
		}
		uniqueSigs[signed.GetAuthor()] = signed.Signature
	}

	quorumCheckpoint := &rbft.QuorumCheckpoint{
		Checkpoint: checkpoint,
		Signatures: lo.MapEntries(uniqueSigs, func(signer ValidatorID, sign Signature) (uint64, []byte) {
			return signer, sign
		}),
	}

	a.stableHeight = height

	stableBlock, err := a.config.GetBlockFunc(height)
	if err != nil {
		a.logger.Errorf("failed to get stable block: %d", height)
		return
	}

	a.postAttestationEvent(stableBlock, quorumCheckpoint)

}

func (a *RBFTAdaptor) postAttestationEvent(block *types.Block, ckp *rbft.QuorumCheckpoint) {
	defer metrics.AttestationCounter.WithLabelValues(consensustypes.Rbft).Inc()
	ckpt := &rbft.RbftQuorumCheckpoint{
		QuorumCheckpoint: ckp,
	}

	proofData, err := ckpt.Marshal()
	if err != nil {
		a.logger.Error(err)
		return
	}

	blockBytes, err := block.Marshal()
	if err != nil {
		a.logger.Error(err)
		return
	}
	a.AttestationFeed.Send(events.AttestationEvent{
		AttestationData: &consensustypes.Attestation{
			Epoch:         block.Header.Epoch,
			ConsensusType: repo.ConsensusTypeRbft,
			Block:         blockBytes,
			Proof:         proofData,
		}})
	if err = a.epochManager.StoreLatestProof(proofData); err != nil {
		a.logger.Error(err)
		return
	}
}
