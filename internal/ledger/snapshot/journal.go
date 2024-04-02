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
	return snap.blockjournal.GetJournalRange()
}

func (snap *Snapshot) GetBlockJournal(height uint64) (*types.SnapshotJournal, error) {
	data, err := snap.blockjournal.Retrieve(height)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	journal, err := types.DecodeSnapshotJournal(data)
	if err != nil {
		panic(err)
	}

	return journal, nil
}
