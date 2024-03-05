package snapshot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
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

func (entry *BlockJournalEntry) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[BlockJournalEntry]: addr=%v,prevAccount=%v,changed=%v,prevStates=[",
		entry.Address.String(), entry.PrevAccount, entry.AccountChanged))
	for k, v := range entry.PrevStates {
		builder.WriteString(fmt.Sprintf("k=%v,v=%v;", hexutil.Encode([]byte(k)), hexutil.Encode(v)))
	}
	builder.WriteString("]")
	return builder.String()
}

func revertJournal(journal *BlockJournalEntry, batch storage.Batch) {
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
		decodedKey, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			panic(fmt.Sprintf("decode key from base64 err: %v", err))
		}
		if val != nil {
			batch.Put(utils.CompositeStorageKey(journal.Address, decodedKey), val)
		} else {
			batch.Delete(utils.CompositeStorageKey(journal.Address, decodedKey))
		}
	}
}

func (snap *Snapshot) GetJournalRange() (uint64, uint64) {
	minHeight := uint64(0)
	maxHeight := uint64(0)

	data := snap.snapStorage.Get(utils.CompositeKey(utils.SnapshotKey, utils.MinHeightStr))
	if data != nil {
		minHeight = utils.UnmarshalHeight(data)
	}

	data = snap.snapStorage.Get(utils.CompositeKey(utils.SnapshotKey, utils.MaxHeightStr))
	if data != nil {
		maxHeight = utils.UnmarshalHeight(data)
	}

	return minHeight, maxHeight
}

func (snap *Snapshot) GetBlockJournal(height uint64) *BlockJournal {
	data := snap.snapStorage.Get(utils.CompositeKey(utils.SnapshotKey, height))
	if data == nil {
		return nil
	}

	journal := &BlockJournal{}
	if err := json.Unmarshal(data, journal); err != nil {
		panic(err)
	}

	return journal
}
