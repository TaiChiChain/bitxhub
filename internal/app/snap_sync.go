package app

import (
	"fmt"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	common2 "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/sirupsen/logrus"
)

type snapMeta struct {
	snapBlock        *types.Block
	snapPersistEpoch uint64
	snapPeers        []string
}

func loadSnapMeta(lg *ledger.Ledger, rep *repo.Repo) (*snapMeta, error) {
	meta, err := lg.StateLedger.GetTrieSnapshotMeta()
	if err != nil {
		return nil, fmt.Errorf("get snapshot meta hash: %w", err)
	}

	snapPersistedEpoch := meta.EpochInfo.Epoch - 1
	if meta.EpochInfo.EpochPeriod+meta.EpochInfo.StartBlock-1 == meta.Block.Height() {
		snapPersistedEpoch = meta.EpochInfo.Epoch
	}

	var peers []string

	// get the validator set of the current local epoch
	for _, v := range meta.EpochInfo.ValidatorSet {
		if v.P2PNodeID != rep.P2PID {
			peers = append(peers, v.P2PNodeID)
		}
	}

	return &snapMeta{
		snapBlock:        meta.Block,
		snapPeers:        peers,
		snapPersistEpoch: snapPersistedEpoch,
	}, nil
}

func (axm *AxiomLedger) prepareSnapSync(latestHeight uint64) error {
	err := axm.Sync.SwitchMode(common.SyncModeSnapshot)
	if err != nil {
		return fmt.Errorf("switch mode err: %w", err)
	}

	var startEpcNum uint64 = 1

	if latestHeight != 0 {
		block, err := axm.ViewLedger.ChainLedger.GetBlock(latestHeight)
		if err != nil {
			return fmt.Errorf("get latest block err: %w", err)
		}
		blockEpc := block.BlockHeader.Epoch
		// todo: wait fix snap store (system contract state)
		info, err := base.GetEpochInfo(axm.ViewLedger.StateLedger, blockEpc)
		if err != nil {
			return fmt.Errorf("get epoch info err: %w", err)
		}
		if info.StartBlock+info.EpochPeriod-1 == latestHeight {
			// if the last block in this epoch had been persisted, start from the next epoch
			startEpcNum = info.Epoch + 1
		} else {
			startEpcNum = info.Epoch
		}
	}

	opts := []common.Option{
		common.WithPeers(axm.snapMeta.snapPeers),
		common.WithStartEpochChangeNum(startEpcNum),
		common.WithLatestPersistEpoch(rbft.GetLatestEpochQuorumCheckpoint(axm.epochStore.Get)),
		common.WithSnapCurrentEpoch(axm.snapMeta.snapPersistEpoch),
	}

	res, err := axm.Sync.Prepare(opts...)
	if err != nil {
		return fmt.Errorf("prepare sync err: %w", err)
	}

	// start chain data sync
	snapCheckpoint := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			Epoch: axm.snapMeta.snapPersistEpoch,
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: axm.snapMeta.snapBlock.Height(),
				Digest: axm.snapMeta.snapBlock.BlockHash.String(),
			},
		},
	}

	err = axm.startSnapSync(snapCheckpoint, axm.snapMeta.snapPeers, latestHeight+1, res.Data.([]*consensus.EpochChange))
	if err != nil {
		return fmt.Errorf("start snap sync err: %w", err)
	}

	err = axm.Sync.SwitchMode(common.SyncModeFull)
	if err != nil {
		return fmt.Errorf("switch mode err: %w", err)
	}

	return nil
}

func (axm *AxiomLedger) startSnapSync(ckpt *consensus.SignedCheckpoint, peers []string, startHeight uint64, epochChanges []*consensus.EpochChange) error {
	syncTaskDoneCh := make(chan error, 1)
	targetHeight := ckpt.Height()
	params := axm.genSnapSyncParams(peers, startHeight, targetHeight, ckpt, epochChanges)
	if err := axm.Sync.StartSync(params, syncTaskDoneCh); err != nil {
		return err
	}

	for {
		select {
		case <-axm.Ctx.Done():
			return nil
		case err := <-syncTaskDoneCh:
			if err != nil {
				return err
			}
		case data := <-axm.Sync.Commit():
			snapData, ok := data.(*common.SnapCommitData)
			if !ok {
				return fmt.Errorf("invalid commit data type: %T", data)
			}

			err := axm.persistChainData(snapData)
			if err != nil {
				return err
			}

			if snapData.Data[len(snapData.Data)-1].GetHeight() == targetHeight {
				axm.logger.WithFields(logrus.Fields{
					"targetHeight": targetHeight,
				}).Info("snap sync task done")
				return nil
			}
		}
	}
}

func (axm *AxiomLedger) persistChainData(data *common.SnapCommitData) error {
	for _, commitData := range data.Data {
		chainData := commitData.(*common.ChainData)
		if err := axm.ViewLedger.ChainLedger.PersistExecutionResult(chainData.Block, chainData.Receipts); err != nil {
			return err
		}
	}

	storeEpochStateFn := func(key string, value []byte) error {
		return common2.StoreEpochState(axm.epochStore, key, value)
	}
	if data.EpochState != nil {
		if err := rbft.PersistEpochQuorumCheckpoint(storeEpochStateFn, data.EpochState); err != nil {
			return err
		}
	}
	return nil
}

func (axm *AxiomLedger) genSnapSyncParams(peers []string, startHeight, targetHeight uint64,
	quorumCkpt *consensus.SignedCheckpoint, epochChanges []*consensus.EpochChange) *common.SyncParams {

	latestBlockHash := axm.ViewLedger.ChainLedger.GetChainMeta().BlockHash.String()
	return &common.SyncParams{
		Peers:            peers,
		LatestBlockHash:  latestBlockHash,
		Quorum:           common2.GetQuorum(axm.Repo.Config.Consensus.Type, len(peers)),
		CurHeight:        startHeight,
		TargetHeight:     targetHeight,
		QuorumCheckpoint: quorumCkpt,
		EpochChanges:     epochChanges,
	}

}
