package snapshot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/axiomesh/axiom-kit/hexutil"
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
		batch.Put(CompositeSnapAccountKey(journal.Address.String()), data)
	} else {
		batch.Delete(CompositeSnapAccountKey(journal.Address.String()))
	}

	for key, val := range journal.PrevStates {
		decodedKey, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			panic(fmt.Sprintf("decode key from base64 err: %v", err))
		}
		if val != nil {
			batch.Put(CompositeSnapStorageKey(journal.Address.String(), decodedKey), val)
		} else {
			batch.Delete(CompositeSnapStorageKey(journal.Address.String(), decodedKey))
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
