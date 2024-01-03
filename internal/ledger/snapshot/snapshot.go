package snapshot

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
)

type Snapshot struct {
	origin Layer // disklayer

	// todo multi diff layers

	logger logrus.FieldLogger
	diskdb storage.Storage
}

var (
	ErrorRollbackToHigherNumber  = errors.New("rollback to higher blockchain height")
	ErrorRollbackTooMuch         = errors.New("rollback too much block")
	ErrorRemoveJournalOutOfRange = errors.New("remove journal out of range")
	ErrorTargetLayerNotFound     = errors.New("can not find target layer")
)

func NewSnapshot(diskdb storage.Storage, logger logrus.FieldLogger) *Snapshot {
	return &Snapshot{
		diskdb: diskdb,
		logger: logger,
		origin: NewDiskLayer(diskdb),
	}
}

// RemoveJournalsBeforeBlock removes snapshot journals whose block number < height
func (snap *Snapshot) RemoveJournalsBeforeBlock(height uint64) error {
	minHeight, maxHeight := GetJournalRange(snap.diskdb)
	if height > maxHeight {
		return ErrorRemoveJournalOutOfRange
	}

	if height <= minHeight {
		return nil
	}

	batch := snap.diskdb.NewBatch()
	for i := minHeight; i < height; i++ {
		batch.Delete(CompositeSnapJournalKey(i))
	}
	batch.Put(CompositeSnapJournalKey(MinHeightStr), marshalHeight(height))
	batch.Commit()

	return nil
}

func (snap *Snapshot) Account(addr *types.Address) (*types.InnerAccount, error) {
	if snap.origin == nil {
		return nil, ErrorTargetLayerNotFound
	}
	return snap.origin.Account(addr)
}

func (snap *Snapshot) Storage(addr *types.Address, key []byte) ([]byte, error) {
	if snap.origin == nil {
		return nil, ErrorTargetLayerNotFound
	}
	return snap.origin.Storage(addr, key)
}

func (snap *Snapshot) Update(stateRoot common.Hash, destructs map[string]struct{}, accounts map[string]*types.InnerAccount, storage map[string]map[string][]byte) error {
	if snap.origin == nil {
		return ErrorTargetLayerNotFound
	}
	snap.origin.Update(stateRoot, destructs, accounts, storage)
	return nil
}

func (snap *Snapshot) UpdateJournal(height uint64, journal *BlockJournal) error {
	snap.logger.Infof("[UpdateJournal] height:%v", height)

	for _, entry := range journal.Journals {
		entry.PrevStates = encodeToBase64(entry.PrevStates)
	}

	data, err := json.Marshal(journal)
	if err != nil {
		return fmt.Errorf("marshal snapshot journal error: %w", err)
	}
	batch := snap.diskdb.NewBatch()
	batch.Put(CompositeSnapJournalKey(height), data)
	batch.Put(CompositeSnapJournalKey(MaxHeightStr), marshalHeight(height))
	if height == 1 {
		batch.Put(CompositeSnapJournalKey(MinHeightStr), marshalHeight(height))
	}
	batch.Commit()
	return nil
}

// Rollback removes snapshot journals whose block number < height
func (snap *Snapshot) Rollback(height uint64) error {
	minHeight, maxHeight := GetJournalRange(snap.diskdb)
	snap.logger.Infof("[Snapshot-Rollback],minHeight=%v,maxHeight=%v,height=%v", minHeight, maxHeight, height)

	if maxHeight < height {
		return ErrorRollbackToHigherNumber
	}

	if minHeight > height && !(minHeight == 1 && height == 0) {
		return ErrorRollbackTooMuch
	}

	if maxHeight == height {
		return nil
	}

	for i := maxHeight; i > height; i-- {
		snap.logger.Debugf("[Snapshot-Rollback] execute journal of height %v", i)
		batch := snap.diskdb.NewBatch()
		blockJournal := GetBlockJournal(i, snap.diskdb)
		if blockJournal == nil {
			return ErrorRemoveJournalOutOfRange
		}
		for _, entry := range blockJournal.Journals {
			snap.logger.Debugf("[Snapshot-Rollback] execute entry: %v", entry.String())
			revertJournal(entry, batch)
		}
		batch.Delete(CompositeSnapJournalKey(i))
		batch.Put(CompositeSnapJournalKey(MaxHeightStr), marshalHeight(i-1))
		batch.Commit()
	}

	snap.origin.Clear()

	return nil
}

func encodeToBase64(src map[string][]byte) map[string][]byte {
	encodedData := make(map[string][]byte)
	for key, value := range src {
		encodedKey := base64.StdEncoding.EncodeToString([]byte(key))
		encodedData[encodedKey] = value
	}
	return encodedData
}
