package snapshot

import (
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
)

func revertJournal(journal *types.SnapshotJournalEntry, batch kv.Batch) {
	if !journal.AccountChanged {
		return
	}
	if journal.PrevAccount != nil {
		data, err := journal.PrevAccount.Marshal()
		if err != nil {
			panic(err)
		}
		batch.Put(utils.CompositeAccountKey(journal.Address), data)
	} else {
		batch.Delete(utils.CompositeAccountKey(journal.Address))
	}

	for key, val := range journal.PrevStates {
		if len(val) > 0 {
			batch.Put(utils.CompositeStorageKey(journal.Address, []byte(key)), val)
		} else {
			batch.Delete(utils.CompositeStorageKey(journal.Address, []byte(key)))
		}
	}
}

func (snap *Snapshot) GetJournalRange() (uint64, uint64) {
	minHeight := uint64(0)
	maxHeight := uint64(0)

	data := snap.backend.Get(utils.CompositeKey(utils.SnapshotKey, utils.MinHeightStr))
	if data != nil {
		minHeight = utils.UnmarshalUint64(data)
	}

	data = snap.backend.Get(utils.CompositeKey(utils.SnapshotKey, utils.MaxHeightStr))
	if data != nil {
		maxHeight = utils.UnmarshalUint64(data)
	}

	return minHeight, maxHeight
}

func (snap *Snapshot) GetBlockJournal(height uint64) *types.SnapshotJournal {
	data := snap.backend.Get(utils.CompositeKey(utils.SnapshotKey, height))
	if data == nil {
		return nil
	}

	journal, err := types.DecodeSnapshotJournal(data)
	if err != nil {
		panic(err)
	}

	return journal
}
