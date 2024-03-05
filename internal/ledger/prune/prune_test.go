package prune

import (
	"path/filepath"
	"testing"

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

func createMockRepo(t *testing.T) *repo.Repo {
	r, err := repo.Default(t.TempDir())
	require.Nil(t, err)
	r.Config.Ledger.StateLedgerHistoryBlockNumber = 10
	return r
}
