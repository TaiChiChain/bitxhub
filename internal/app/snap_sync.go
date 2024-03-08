package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	common2 "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type snapMeta struct {
	snapBlockHeader  *types.BlockHeader
	snapPersistEpoch uint64
	snapPeers        []string
}

func loadSnapMeta(lg *ledger.Ledger, rep *repo.Repo) (*snapMeta, error) {
	meta, err := lg.StateLedger.GetTrieSnapshotMeta()
	if err != nil {
		return nil, fmt.Errorf("get snapshot meta hash: %w", err)
	}

	snapPersistedEpoch := meta.EpochInfo.Epoch - 1
	if meta.EpochInfo.EpochPeriod+meta.EpochInfo.StartBlock-1 == meta.BlockHeader.Number {
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
		snapBlockHeader:  meta.BlockHeader,
		snapPeers:        peers,
		snapPersistEpoch: snapPersistedEpoch,
	}, nil
}

func (axm *AxiomLedger) prepareSnapSync(latestHeight uint64) (*common.PrepareData, *consensus.SignedCheckpoint, error) {
	// 1. switch to snapshot mode
	err := axm.Sync.SwitchMode(common.SyncModeSnapshot)
	if err != nil {
		return nil, nil, fmt.Errorf("switch mode err: %w", err)
	}

	var startEpcNum uint64 = 1

	if latestHeight != 0 {
		blockHeader, err := axm.ViewLedger.ChainLedger.GetBlockHeader(latestHeight)
		if err != nil {
			return nil, nil, fmt.Errorf("get latest blockHeader err: %w", err)
		}
		blockEpc := blockHeader.Epoch
		info, err := base.GetEpochInfo(axm.ViewLedger.StateLedger, blockEpc)
		if err != nil {
			return nil, nil, fmt.Errorf("get epoch info err: %w", err)
		}
		if info.StartBlock+info.EpochPeriod-1 == latestHeight {
			// if the last blockHeader in this epoch had been persisted, start from the next epoch
			startEpcNum = info.Epoch + 1
		} else {
			startEpcNum = info.Epoch
		}
	}

	// 2. fill snap sync config option
	opts := []common.Option{
		common.WithPeers(axm.snapMeta.snapPeers),
		common.WithStartEpochChangeNum(startEpcNum),
		common.WithLatestPersistEpoch(rbft.GetLatestEpochQuorumCheckpoint(axm.epochStore.Get)),
		common.WithSnapCurrentEpoch(axm.snapMeta.snapPersistEpoch),
	}

	// 3. prepare snap sync info
	res, err := axm.Sync.Prepare(opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("prepare sync err: %w", err)
	}

	snapCheckpoint := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			Epoch: axm.snapMeta.snapPersistEpoch,
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: axm.snapMeta.snapBlockHeader.Number,
				Digest: axm.snapMeta.snapBlockHeader.Hash().String(),
			},
		},
	}

	return res, snapCheckpoint, nil
}

func (axm *AxiomLedger) startSnapSync(verifySnapCh chan bool, ckpt *consensus.SignedCheckpoint, peers []string, startHeight uint64, epochChanges []*consensus.EpochChange) error {
	syncTaskDoneCh := make(chan error, 1)
	targetHeight := ckpt.Height()
	params := axm.genSnapSyncParams(peers, startHeight, targetHeight, ckpt, epochChanges)
	start := time.Now()
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
			now := time.Now()
			snapData, ok := data.(*common.SnapCommitData)
			if !ok {
				return fmt.Errorf("invalid commit data type: %T", data)
			}

			err := axm.persistChainData(snapData)
			if err != nil {
				return err
			}
			currentHeight := snapData.Data[len(snapData.Data)-1].GetHeight()
			axm.logger.WithFields(logrus.Fields{
				"Height": currentHeight,
				"target": targetHeight,
				"cost":   time.Since(now),
			}).Info("persist chain data task")

			if currentHeight == targetHeight {
				axm.logger.WithFields(logrus.Fields{
					"targetHeight": targetHeight,
					"cost":         time.Since(start),
				}).Info("snap sync task done")

				if !axm.waitVerifySnapTrie(verifySnapCh) {
					return errors.New("verify snap trie failed")
				}
				return nil
			}
		}
	}
}

func (axm *AxiomLedger) waitVerifySnapTrie(verifySnapCh chan bool) bool {
	return <-verifySnapCh
}

func (axm *AxiomLedger) persistChainData(data *common.SnapCommitData) error {
	var batchBlock []*types.Block
	var batchReceipts [][]*types.Receipt
	for _, commitData := range data.Data {
		chainData := commitData.(*common.ChainData)
		batchBlock = append(batchBlock, chainData.Block)
		batchReceipts = append(batchReceipts, chainData.Receipts)
	}
	if len(batchBlock) > 0 {
		if err := axm.ViewLedger.ChainLedger.BatchPersistExecutionResult(batchBlock, batchReceipts); err != nil {
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
