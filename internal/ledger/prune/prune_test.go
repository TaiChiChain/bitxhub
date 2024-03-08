package prune

import (
	"crypto/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestTrieCacheUpdate(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	tc := NewTrieCache(createMockRepo(t), pStateStorage, logger)
	v, ok := tc.Get(0, []byte("empty"))
	require.False(t, ok)
	require.Nil(t, v)

	trieJournal1 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string][]byte{
				"k1":                       []byte("v1"),
				"k2":                       []byte("v2"),
				string([]byte{0, 0, 0, 1}): []byte("vv"),
			},
			PruneSet: map[string]struct{}{},
		},
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string][]byte{
				"k3": []byte("v3"),
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
			DirtySet: map[string][]byte{
				"k1": []byte("v11"),
				"k5": []byte("v55"),
			},
			PruneSet: map[string]struct{}{
				"k2": {},
			},
		},
		&types.TrieJournal{
			Version:  2,
			DirtySet: map[string][]byte{},
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
	require.Equal(t, v, []byte("v1"))
	v, ok = tc.Get(1, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v3"))
	v, ok = tc.Get(1, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(1, []byte("k5"))
	require.False(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(1, []byte{0, 0, 0, 1})
	require.True(t, ok)
	require.Equal(t, v, []byte("vv"))

	// verify version 2
	v, ok = tc.Get(2, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v11"))
	v, ok = tc.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k5"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v55"))

}

func TestTrieCacheRollback(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	tc := NewTrieCache(createMockRepo(t), pStateStorage, logger)

	trieJournal1 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string][]byte{
				"k1": []byte("v1"),
				"k2": []byte("v2"),
			},
			PruneSet: map[string]struct{}{},
		},
		&types.TrieJournal{
			Version: 1,
			DirtySet: map[string][]byte{
				"k3": []byte("v3"),
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
			DirtySet: map[string][]byte{
				"k1": []byte("v11"),
				"k5": []byte("v55"),
			},
			PruneSet: map[string]struct{}{
				"k2": {},
			},
		},
		&types.TrieJournal{
			Version:  2,
			DirtySet: map[string][]byte{},
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
	require.Equal(t, v, []byte("v1"))
	v, ok = tc.Get(1, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v3"))
	v, ok = tc.Get(1, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(1, []byte("k5"))
	require.False(t, ok)
	require.Nil(t, v)

	// verify version 2
	v, ok = tc.Get(2, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v1"))
	v, ok = tc.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v3"))
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
			DirtySet: map[string][]byte{
				"k1": []byte("v11"),
				"k5": []byte("v55"),
			},
			PruneSet: map[string]struct{}{
				"k2": {},
			},
		},
		&types.TrieJournal{
			Version:  2,
			DirtySet: map[string][]byte{},
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
	require.Equal(t, v, []byte("v11"))
	v, ok = tc.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = tc.Get(2, []byte("k5"))
	require.True(t, ok)
	require.Equal(t, v, []byte("v55"))

}

func TestPruningFlushByMaxFlushNum(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	tc := NewTrieCache(createMockRepo(t), pStateStorage, logger)
	for i := 0; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1; i++ {
		trieJournal1 := types.TrieJournalBatch{
			&types.TrieJournal{
				Version: uint64(i + 1),
				DirtySet: map[string][]byte{
					"k1": []byte("v1"),
				},
				PruneSet: map[string]struct{}{},
			},
		}
		tc.Update(uint64(i+1), trieJournal1)
	}
	time.Sleep(time.Second)
	v, ok := tc.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, []byte("v1"), v)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, []byte("v1"), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	trieJournal2 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 2,
			DirtySet: map[string][]byte{
				"k1": []byte("v11"),
			},
		},
	}
	tc.Update(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), trieJournal2)
	time.Sleep(checkFlushTimeInterval + time.Second)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, []byte("v11"), v)
	require.Equal(t, []byte("v1"), tc.ledgerStorage.Get([]byte("k1"))) // backend flushed
	require.Equal(t, tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(tc.states.diffs))

}

func TestPruningFlushByMaxBatchSize(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	tc := NewTrieCache(createMockRepo(t), pStateStorage, logger)
	bigV := make([]byte, maxFlushBatchSizeThreshold)
	rand.Read(bigV)
	trieJournal := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: uint64(1),
			DirtySet: map[string][]byte{
				"k1": bigV,
			},
		},
	}
	tc.Update(uint64(1), trieJournal)

	for i := 1; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum; i++ {
		trieJournal := types.TrieJournalBatch{
			&types.TrieJournal{
				Version: uint64(i + 1),
				DirtySet: map[string][]byte{
					"k1": []byte("v1"),
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
	require.Equal(t, []byte("v1"), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	trieJournal2 := types.TrieJournalBatch{
		&types.TrieJournal{
			Version: 2,
			DirtySet: map[string][]byte{
				"k1": []byte("v11"),
			},
		},
	}
	tc.Update(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), trieJournal2)
	time.Sleep(2*checkFlushTimeInterval + time.Second)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, []byte("v11"), v)
	v = tc.ledgerStorage.Get([]byte("k1"))
	require.True(t, len(v) > 0)
	require.Equal(t, bigV, v) // flushed
	require.Equal(t, tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(tc.states.diffs))

}

func TestPruningFlushByMaxFlushTime(t *testing.T) {

	t.Skip()

	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	tc := NewTrieCache(createMockRepo(t), pStateStorage, logger)

	for i := 0; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1; i++ {
		trieJournal := types.TrieJournalBatch{
			&types.TrieJournal{
				Version: uint64(i + 1),
				DirtySet: map[string][]byte{
					"k1": []byte("v1"),
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
	require.Equal(t, []byte("v1"), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	time.Sleep(maxFlushTimeInterval + time.Second)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, []byte("v1"), v)
	v = tc.ledgerStorage.Get([]byte("k1"))
	require.True(t, len(v) > 0)
	require.Equal(t, []byte("v1"), v) // flushed
	require.Equal(t, tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(tc.states.diffs))

}

func createMockRepo(t *testing.T) *repo.Repo {
	r, err := repo.Default(t.TempDir())
	require.Nil(t, err)
	r.Config.Ledger.StateLedgerReservedHistoryBlockNum = 10
	return r
}
