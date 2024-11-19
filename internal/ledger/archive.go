package ledger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type Archiver struct {
	rep *repo.Repo

	chainState *chainstate.ChainState
	//archiveHistoryBackend      kv.Storage
	//archiveJournalBackend      kv.Storage
	archiveTrieSnapshotBackend kv.Storage

	archiveTrieSnapshotPath string

	archiveStateLedger *ArchiveStateLedger

	ledgerBackend   kv.Storage
	snapshotBackend kv.Storage

	logger logrus.FieldLogger
}

type ArchiveArgs struct {
	ArchiveHistoryStorage kv.Storage
	ArchiveJournalStorage kv.Storage
}

func NewArchiver(rep *repo.Repo, archiveArgs *ArchiveArgs, logger logrus.FieldLogger) *Archiver {
	snapshotPath := storagemgr.GetLedgerComponentPath(rep, storagemgr.ArchiveSnapshot)
	archiveSnapshotStorage, err := storagemgr.Open(snapshotPath)
	if err != nil {
		panic(err)
	}
	archiveStateLedger := NewArchiveStateLedger(rep, 0)
	archiver := &Archiver{
		rep:                        rep,
		archiveTrieSnapshotBackend: archiveSnapshotStorage,
		logger:                     logger,
		archiveTrieSnapshotPath:    snapshotPath,
		archiveStateLedger:         archiveStateLedger,
	}
	return archiver
}

func (archiver *Archiver) Archive(blockHeader *types.BlockHeader, stateJournal *types.StateJournal) (err error) {
	if blockHeader.Number != 0 && (archiver.chainState == nil || !archiver.chainState.IsDataSyncer) {
		archiver.logger.Infof("[Archive] current node doesn't support archive, skip it")
		return nil
	}

	cur := time.Now()
	var wg sync.WaitGroup

	// 1. archive history data and journal data in rust ledger
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = archiver.archiveStateLedger.Archive(blockHeader, stateJournal)
	}()

	// 2. update trie snapshot data
	wg.Add(1)
	go func() {
		defer wg.Done()
		snapshotBatch := archiver.archiveTrieSnapshotBackend.NewBatch()

		// 2.1 apply state diff
		for _, journal := range stateJournal.TrieJournal {
			snapshotBatch.Put(journal.RootHash[:], journal.RootNodeKey.Encode())
			for k, v := range journal.DirtySet {
				snapshotBatch.Put([]byte(k), v.Encode())
			}
			for k := range journal.PruneSet {
				snapshotBatch.Delete([]byte(k))
			}
		}
		for k, v := range stateJournal.CodeJournal {
			snapshotBatch.Put([]byte(k), v)
		}

		// 2.2 update meta info
		epochInfo := archiver.rep.GenesisConfig.EpochInfo
		if blockHeader.Number != 0 {
			if epochInfo, err = archiver.chainState.GetEpochInfo(blockHeader.Epoch); err != nil {
				panic(fmt.Errorf("get epoch info failed: %w", err))
			}
		}
		snapshotMeta := &utils.SnapshotMeta{
			BlockHeader: blockHeader,
			EpochInfo:   epochInfo,
		}
		snapshotMetaBytes, err := snapshotMeta.Marshal()
		if err != nil {
			panic(fmt.Errorf("marshal snapshotMeta failed: %w", err))
		}
		snapshotBatch.Put([]byte(utils.SnapshotMetaKey), snapshotMetaBytes)
		snapshotBatch.Put(utils.CompositeKey(utils.PruneJournalKey, utils.MinHeightStr), utils.MarshalUint64(snapshotMeta.BlockHeader.Number))
		snapshotBatch.Put(utils.CompositeKey(utils.PruneJournalKey, utils.MaxHeightStr), utils.MarshalUint64(snapshotMeta.BlockHeader.Number))

		snapshotBatch.Commit()
	}()

	wg.Wait()

	archiver.logger.Infof("[Archive] archive history at height: %v, time: %v", blockHeader.Number, time.Since(cur))
	return nil
}

func (archiver *Archiver) NewView(blockHeader *types.BlockHeader) (*ArchiveStateLedger, error) {
	res, err := archiver.archiveStateLedger.NewView(blockHeader, false)
	return res.(*ArchiveStateLedger), err
}

func (archiver *Archiver) ExportArchivedSnapshot(targetFilePath string) error {
	panic("rust archive ledger don't support ExportArchivedSnapshot yet")
	//blockHeader := &types.BlockHeader{}
	//blockHeaderBlob := archiver.archiveJournalBackend.Get(utils.CompositeKey(utils.ArchiveKey, utils.BlockHeader))
	//if err := blockHeader.Unmarshal(blockHeaderBlob); err != nil {
	//	return err
	//}
	//
	//cur := time.Now()
	//archiver.logger.Infof("[ExportArchivedSnapshot] start archive snapshot at height: %v", blockHeader.Number)
	//
	//if err := archiver.archiveTrieSnapshotBackend.Close(); err != nil {
	//	return errors.Errorf("archive snapshot error: %v", err)
	//}
	//snapshotTargetPath := filepath.Join(targetFilePath, fmt.Sprintf("snapshot-%v-%v", blockHeader.Number, time.Now().Format("2006-01-02T15-04-05")))
	//if err := os.MkdirAll(snapshotTargetPath, os.ModePerm); err != nil {
	//	return errors.Errorf("mkdir snapshot archive dir error: %v", err.Error())
	//}
	//if err := copyDir(archiver.archiveTrieSnapshotPath, snapshotTargetPath); err != nil {
	//	return errors.Errorf("copy archived snapshot error: %v", err)
	//}
	//
	//archiver.logger.Infof("[ExportArchivedSnapshot] finish archive snapshot at height: %v, time: %v", blockHeader.Number, time.Since(cur))
	//return nil
}

func (archiver *Archiver) UpdateChainState(chainState *chainstate.ChainState) {
	archiver.chainState = chainState
}

//func (archiver *Archiver) GetHistoryBackend() kv.Storage {
//	return archiver.archiveHistoryBackend
//}

func (archiver *Archiver) GetStateJournal(height uint64) *types.StateJournal {
	return archiver.archiveStateLedger.GetStateJournal(height)
}

func copyDir(src, dest string) error {
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		destPath := filepath.Join(dest, file.Name())

		if file.IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return errors.Errorf("mkdir %s dir error: %v", destPath, err.Error())
			}
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
