package trie_indexer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNormalCase(t *testing.T) {
	logger := log.NewWithModule("trie_indexer_test")
	pStateStorage := kv.NewMemory()

	indexer := NewTrieIndexer(createMockRepo(t), pStateStorage, logger)

	stateDelta := &types.StateDelta{
		Journal: make([]*types.TrieJournal, 1),
	}

	addr := leftPadBytes([]byte{'a'}, 40)

	internalNk := &types.NodeKey{
		Path:    []byte{'1', '2', 'a', 'b'},
		Version: 1,
		Type:    addr,
	}

	leafNk := &types.NodeKey{
		Path:    []byte{'a', 'b'},
		Version: 1,
		Type:    addr,
	}

	internalNode := &types.InternalNode{}
	originKey := []byte{'a', 'b', 'c'}
	leafKey := utils.CompositeKey(utils.TrieNodeIndexKey, originKey)
	leafNode := &types.LeafNode{Key: leafKey, Val: []byte("bbb")}

	stateDelta.Journal[0] = &types.TrieJournal{
		DirtySet: map[string]types.Node{
			string(internalNk.Encode()): internalNode,
			string(leafNk.Encode()):     leafNode,
		},
	}

	indexer.Update(1, stateDelta)

	res := indexer.GetTrieIndexes(1, addr, originKey)
	require.NotNil(t, res)
}

func createMockRepo(t *testing.T) *repo.Repo {
	r := repo.MockRepo(t)
	return r
}

func leftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}
