package prune

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
)

type diffLayer struct {
	height uint64

	cache map[string][]byte

	// todo implement here in a more elegant way
	accountCache map[string][]byte
	storageCache map[string][]byte

	ledgerStorage storage.Storage
}

const TypeAccount = 1
const TypeStorage = 2

func NewDiffLayer(height uint64, ledgerStorage storage.Storage, trieJournals types.TrieJournalBatch, persist bool) *diffLayer {
	l := &diffLayer{
		height:        height,
		ledgerStorage: ledgerStorage,
		cache:         make(map[string][]byte),
		accountCache:  make(map[string][]byte),
		storageCache:  make(map[string][]byte),
	}
	if persist {
		batch := l.ledgerStorage.NewBatch()
		batch.Put(utils.CompositeKey(utils.TrieJournalKey, height), trieJournals.Encode())
		batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MaxHeightStr), utils.MarshalHeight(height))
		if height == 1 {
			batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr), utils.MarshalHeight(height))
		}
		batch.Commit()
	}

	for _, journal := range trieJournals {
		for k := range journal.PruneSet {
			l.cache[k] = nil
			if journal.Type == TypeAccount {
				l.accountCache[k] = nil
			} else {
				l.storageCache[k] = nil
			}
		}
		for k, v := range journal.DirtySet {
			l.cache[k] = v
			if journal.Type == TypeAccount {
				l.accountCache[k] = v
			} else {
				l.storageCache[k] = v
			}
		}
	}

	return l
}

func (dl *diffLayer) GetFromTrieCache(nodeKey []byte) ([]byte, bool) {
	if v, ok := dl.cache[string(nodeKey)]; ok {
		return v, true
	}
	return nil, false
}

func (dl *diffLayer) String() string {
	res := strings.Builder{}
	res.WriteString("Version[")
	res.WriteString(strconv.Itoa(int(dl.height)))
	res.WriteString(fmt.Sprintf("], \nJournals[\n"))
	res.WriteString(fmt.Sprintf("journal=%v\n", dl.cache))
	res.WriteString("]")
	return res.String()
}

// TODO: implement count journal size function
