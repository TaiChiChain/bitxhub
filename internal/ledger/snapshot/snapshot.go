package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
)

type Snapshot struct {
	origin Layer // disklayer

	// todo multi diff layers

	diskdb storage.Storage
}

var (
	ErrorRollbackToHigherNumber  = errors.New("rollback to higher blockchain height")
	ErrorRollbackTooMuch         = errors.New("rollback too much block")
	ErrorRemoveJournalOutOfRange = errors.New("remove journal out of range")
	ErrorTargetLayerNotFound     = errors.New("can not find target layer")
)

func NewSnapshot(diskdb storage.Storage) *Snapshot {
	return &Snapshot{
		diskdb: diskdb,
		origin: NewDiskLayer(diskdb),
	}
}

// RemoveJournalsBeforeBlock removes snapshot journals whose block number < height
func (snap *Snapshot) RemoveJournalsBeforeBlock(height uint64) error {
	minHeight, maxHeight := snap.GetJournalRange()
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
	minHeight, maxHeight := snap.GetJournalRange()

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
		batch := snap.diskdb.NewBatch()
		blockJournal := snap.GetBlockJournal(i)
		if blockJournal == nil {
			return ErrorRemoveJournalOutOfRange
		}
		for _, entry := range blockJournal.Journals {
			revertJournal(entry, batch)
		}
		batch.Delete(CompositeSnapJournalKey(i))
		batch.Put(CompositeSnapJournalKey(MaxHeightStr), marshalHeight(i-1))
		batch.Commit()
	}

	snap.origin.Clear()

	return nil
}
