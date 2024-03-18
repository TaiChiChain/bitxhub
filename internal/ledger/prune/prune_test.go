package prune

import (
	"crypto/rand"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func makeLeafNode(str string) *types.LeafNode {
	return &types.LeafNode{
		Val: []byte(str),
	}
}

func TestTrieCacheUpdate(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)
	accountTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	storageTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	tc := NewTrieCache(createMockRepo(t), pStateStorage, accountTrieStorage, storageTrieStorage, logger)
	v, ok := tc.Get(0, []byte("empty"))
	require.False(t, ok)
	require.Nil(t, v)

	trieJournal1 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string]types.Node{
				"k1":                       makeLeafNode("v1"),
				"k2":                       makeLeafNode("v2"),
				string([]byte{0, 0, 0, 1}): makeLeafNode("vv"),
			},
			PruneSet: map[string]struct{}{},
		},
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string]types.Node{
				"k3": makeLeafNode("v3"),
			},
			PruneSet: map[string]struct {
			}{
				"k4": {},
			},
		},
	}
	tc.Update(1, trieJournal1)

	trieJournal2 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 2,
			DirtySet: map[string]types.Node{
				"k1": makeLeafNode("v11"),
				"k5": makeLeafNode("v55"),
			},
			PruneSet: map[string]struct{}{
				"k2": {},
			},
		},
		&types.TrieJournal{
			Version:  2,
			DirtySet: map[string]types.Node{},
			PruneSet: map[string]struct {
			}{
				"k3": {},
			},
		},
	}

	tc.Update(2, trieJournal2)

	// verify version 1
	v, ok = tc.Get(1, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v1"))
	v, ok = tc.Get(1, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v3"))
	v, ok = tc.Get(1, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(1, []byte("k5"))
	require.False(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(1, []byte{0, 0, 0, 1})
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("vv"))

	// verify version 2
	v, ok = tc.Get(2, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v11"))
	v, ok = tc.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k5"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v55"))

}

func TestTrieCacheRollback(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	accountTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	storageTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	tc := NewTrieCache(createMockRepo(t), pStateStorage, accountTrieStorage, storageTrieStorage, logger)

	trieJournal1 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string]types.Node{
				"k1": makeLeafNode("v1"),
				"k2": makeLeafNode("v2"),
			},
			PruneSet: map[string]struct{}{},
		},
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string]types.Node{
				"k3": makeLeafNode("v3"),
			},
			PruneSet: map[string]struct {
			}{
				"k4": {},
			},
		},
	}
	tc.Update(1, trieJournal1)

	trieJournal2 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 2,
			DirtySet: map[string]types.Node{
				"k1": makeLeafNode("v11"),
				"k5": makeLeafNode("v55"),
			},
			PruneSet: map[string]struct{}{
				"k2": {},
			},
		},
		&types.TrieJournal{
			Version:  2,
			DirtySet: map[string]types.Node{},
			PruneSet: map[string]struct {
			}{
				"k3": {},
			},
		},
	}
	tc.Update(2, trieJournal2)

	err = tc.Rollback(1)
	require.Nil(t, err)

	// verify version 1
	v, ok := tc.Get(1, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v1"))
	v, ok = tc.Get(1, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v3"))
	v, ok = tc.Get(1, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(1, []byte("k5"))
	require.False(t, ok)
	require.Nil(t, v)

	// verify version 2
	v, ok = tc.Get(2, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v1"))
	v, ok = tc.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v3"))
	v, ok = tc.Get(2, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k5"))
	require.False(t, ok)
	require.Nil(t, v)

	// replay version 2
	trieJournal2 = types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 2,
			DirtySet: map[string]types.Node{
				"k1": makeLeafNode("v11"),
				"k5": makeLeafNode("v55"),
			},
			PruneSet: map[string]struct{}{
				"k2": {},
			},
		},
		&types.TrieJournal{
			Version:  2,
			DirtySet: map[string]types.Node{},
			PruneSet: map[string]struct {
			}{
				"k3": {},
			},
		},
	}
	tc.Update(2, trieJournal2)

	// verify version 2 again after replay
	v, ok = tc.Get(2, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v11"))
	v, ok = tc.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k5"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v55"))

}

func TestPruningFlushByMaxFlushNum(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	accountTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	storageTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	tc := NewTrieCache(createMockRepo(t), pStateStorage, accountTrieStorage, storageTrieStorage, logger)
	for i := 0; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1; i++ {
		trieJournal1 := types.TrieJournalBatch{
			&types.TrieJournal{
				Version: uint64(i + 1),
				DirtySet: map[string]types.Node{
					"k" + strconv.Itoa(i+1): makeLeafNode("v" + strconv.Itoa(i+1)),
				},
				PruneSet: map[string]struct{}{},
			},
		}
		tc.Update(uint64(i+1), trieJournal1)
	}
	time.Sleep(time.Second)
	v, ok := tc.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1), []byte("k"+strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1)))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v"+strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1)), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	trieJournal2 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum),
			DirtySet: map[string]types.Node{
				"k" + strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum): makeLeafNode("v" + strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum)),
			},
		},
	}
	tc.Update(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), trieJournal2)
	time.Sleep(2*checkFlushTimeInterval + time.Second)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), []byte("k1"))
	require.False(t, ok)
	require.Nil(t, v)
	require.Equal(t, makeLeafNode("v1").EncodePb(), tc.ledgerStorage.Get([]byte("k1"))) // backend flushed
	require.Equal(t, tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(tc.states.diffs))

}

func TestPruningFlushByMaxBatchSize(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	accountTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	storageTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	tc := NewTrieCache(createMockRepo(t), pStateStorage, accountTrieStorage, storageTrieStorage, logger)
	bigV := make([]byte, maxFlushBatchSizeThreshold)
	rand.Read(bigV)
	trieJournal := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: uint64(1),
			DirtySet: map[string]types.Node{
				"k1": makeLeafNode(string(bigV)),
			},
		},
	}
	tc.Update(uint64(1), trieJournal)

	for i := 1; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum; i++ {
		trieJournal := types.TrieJournalBatch{
			&types.TrieJournal{
				Version: uint64(i + 1),
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v1"),
				},
			},
		}
		tc.Update(uint64(i+1), trieJournal)
	}
	time.Sleep(time.Second)
	v, ok := tc.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	trieJournal2 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 2,
			DirtySet: map[string]types.Node{
				"k1": makeLeafNode("v11"),
			},
		},
	}
	tc.Update(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), trieJournal2)
	time.Sleep(2*checkFlushTimeInterval + time.Second)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v11"), v)
	v1 := tc.ledgerStorage.Get([]byte("k1"))
	require.True(t, len(v1) > 0)
	require.Equal(t, makeLeafNode(string(bigV)).EncodePb(), v1) // flushed
	require.Equal(t, tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(tc.states.diffs))

}

func TestPruningFlushByMaxFlushTime(t *testing.T) {

	t.Skip()

	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	accountTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	storageTrieStorage := storagemgr.NewCachedStorage(pStateStorage, 1).(*storagemgr.CachedStorage)
	tc := NewTrieCache(createMockRepo(t), pStateStorage, accountTrieStorage, storageTrieStorage, logger)

	for i := 0; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1; i++ {
		trieJournal := types.TrieJournalBatch{
			&types.TrieJournal{
				Version: uint64(i + 1),
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v1"),
				},
			},
		}
		tc.Update(uint64(i+1), trieJournal)
	}
	time.Sleep(time.Second)
	v, ok := tc.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	time.Sleep(maxFlushTimeInterval + checkFlushTimeInterval + time.Second)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	v1 := tc.ledgerStorage.Get([]byte("k1"))
	require.True(t, len(v1) > 0)
	require.Equal(t, makeLeafNode("v1").EncodePb(), v1) // flushed
	require.Equal(t, 1, len(tc.states.diffs))

}

func TestName(t *testing.T) {
	var arr []int
	arr = append(arr, 1)
	arr = append(arr, 2)
	println(len(arr[1:1]))
}

func createMockRepo(t *testing.T) *repo.Repo {
	r, err := repo.Default(t.TempDir())
	require.Nil(t, err)
	r.Config.Ledger.StateLedgerReservedHistoryBlockNum = 10
	return r
}
