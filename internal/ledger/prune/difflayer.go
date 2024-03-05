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

	trieJournals types.TrieJournalBatch

	ledgerStorage storage.Storage
}

func NewDiffLayer(height uint64, ledgerStorage storage.Storage, trieJournals types.TrieJournalBatch) *diffLayer {
	l := &diffLayer{
		height:        height,
		ledgerStorage: ledgerStorage,
		trieJournals:  trieJournals,
	}
	// persist trie journal and its meta info in disk in case of crash down
	batch := l.ledgerStorage.NewBatch()
	batch.Put(utils.CompositeKey(utils.TrieJournalKey, height), trieJournals.Encode())
	batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MaxHeightStr), utils.MarshalHeight(height))
	if height == 1 {
		batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr), utils.MarshalHeight(height))
	}
	batch.Commit()

	return l
}

func (dl *diffLayer) GetFromTrieCache(nodeKey []byte) (res []byte, ok bool) {
	for _, journal := range dl.trieJournals {
		nk := string(nodeKey)
		if _, ok = journal.PruneSet[nk]; ok {
			break
		}
		if res, ok = journal.DirtySet[nk]; ok {
			break
		}
	}

	//fmt.Printf("[GetFromTrieCache] nodeKey=%v, ok=%v, res=%v,\n", nodeKey, ok, res)
	//fmt.Printf("[GetFromTrieCache] journal is %v,\n", dl.trieJournals)

	return res, ok
}

func (dl *diffLayer) String() string {
	res := strings.Builder{}
	res.WriteString("Version[")
	res.WriteString(strconv.Itoa(int(dl.height)))
	res.WriteString(fmt.Sprintf("], \nJournals[\n"))
	for _, j := range dl.trieJournals {
		res.WriteString(fmt.Sprintf("journal=%v\n", j.String()))
	}
	res.WriteString("]")
	return res.String()
}

// TODO: implement count journal size function
