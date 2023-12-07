package snapshot

import (
	"encoding/json"
	"strconv"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

var (
	MinHeightStr = "minHeight"
	MaxHeightStr = "maxHeight"
)

type BlockJournal struct {
	Journals []*BlockJournalEntry
}

type BlockJournalEntry struct {
	Address        *types.Address
	PrevAccount    *types.InnerAccount
	AccountChanged bool
	PrevStates     map[string][]byte
}

func revertJournal(journal *BlockJournalEntry, batch storage.Batch) {
	if journal.AccountChanged {
		if journal.PrevAccount != nil {
			data, err := journal.PrevAccount.Marshal()
			if err != nil {
				panic(err)
			}
			//fmt.Printf("[revertJournal] journal=%v\n", journal)
			batch.Put(CompositeSnapAccountKey(journal.Address.String()), data)
		} else {
			batch.Delete(CompositeSnapAccountKey(journal.Address.String()))
		}
	}

	for key, val := range journal.PrevStates {
		if val != nil {
			batch.Put(CompositeSnapStorageKey(journal.Address.String(), []byte(key)), val)
		} else {
			batch.Delete(CompositeSnapStorageKey(journal.Address.String(), []byte(key)))
		}
	}
}

func GetJournalRange(ldb storage.Storage) (uint64, uint64) {
	minHeight := uint64(0)
	maxHeight := uint64(0)

	data := ldb.Get(CompositeSnapJournalKey(MinHeightStr))
	if data != nil {
		minHeight = unmarshalHeight(data)
	}

	data = ldb.Get(CompositeSnapJournalKey(MaxHeightStr))
	if data != nil {
		maxHeight = unmarshalHeight(data)
	}

	return minHeight, maxHeight
}

func GetBlockJournal(height uint64, ldb storage.Storage) *BlockJournal {
	data := ldb.Get(CompositeSnapJournalKey(height))
	if data == nil {
		return nil
	}

	journal := &BlockJournal{}
	if err := json.Unmarshal(data, journal); err != nil {
		panic(err)
	}

	return journal
}

func marshalHeight(height uint64) []byte {
	return []byte(strconv.FormatUint(height, 10))
}

func unmarshalHeight(data []byte) uint64 {
	height, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		panic(err)
	}

	return height
}
