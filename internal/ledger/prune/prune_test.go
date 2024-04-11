package prune

import (
	"crypto/rand"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func makeLeafNode(str string) *types.LeafNode {
	return &types.LeafNode{
		Val: []byte(str),
	}
}

func TestPruneCacheUpdate(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil, logrus.New())
	assert.Nil(t, err)
	accountTrieCache := storagemgr.NewCacheWrapper(32, true)
	storageTrieCache := storagemgr.NewCacheWrapper(32, true)
	pruneCache := NewPruneCache(createMockRepo(t), pStateStorage, accountTrieCache, storageTrieCache, logger)
	v, ok := pruneCache.Get(0, []byte("empty"))
	require.False(t, ok)
	require.Nil(t, v)

	batch := pStateStorage.NewBatch()

	trieJournal1 := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef26"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path1"), Type: []byte("type1")},
				DirtySet: map[string]types.Node{
					"k1":                       makeLeafNode("v1"),
					"k2":                       makeLeafNode("v2"),
					string([]byte{0, 0, 0, 1}): makeLeafNode("vv"),
				},
				PruneSet: map[string]struct{}{},
			},
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef27"),
				RootNodeKey: &types.NodeKey{Version: 2, Path: []byte("path2"), Type: []byte("type2")},
				DirtySet: map[string]types.Node{
					"k3": makeLeafNode("v3"),
				},
				PruneSet: map[string]struct {
				}{
					"k4": {},
				},
			},
		},
	}
	pruneCache.Update(batch, 1, trieJournal1)

	trieJournal2 := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 3, Path: []byte("path3"), Type: []byte("type3")},
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v11"),
					"k5": makeLeafNode("v55"),
				},
				PruneSet: map[string]struct{}{
					"k2": {},
				},
			},
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef29"),
				RootNodeKey: &types.NodeKey{Version: 4, Path: []byte("path4"), Type: []byte("type4")},
				DirtySet:    map[string]types.Node{},
				PruneSet: map[string]struct {
				}{
					"k3": {},
				},
			},
		},
	}

	pruneCache.Update(batch, 2, trieJournal2)

	batch.Commit()

	// verify version 1
	v, ok = pruneCache.Get(1, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v1"))
	v, ok = pruneCache.Get(1, []byte("k3"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v3"))
	v, ok = pruneCache.Get(1, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = pruneCache.Get(1, []byte("k5"))
	require.False(t, ok)
	require.Nil(t, v)
	v, ok = pruneCache.Get(1, []byte{0, 0, 0, 1})
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("vv"))

	// verify version 2
	v, ok = pruneCache.Get(2, []byte("k1"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v11"))
	v, ok = pruneCache.Get(2, []byte("k3"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = pruneCache.Get(2, []byte("k4"))
	require.True(t, ok)
	require.Nil(t, v)
	v, ok = pruneCache.Get(2, []byte("k5"))
	require.True(t, ok)
	require.Equal(t, v, makeLeafNode("v55"))

}

func TestPruneCacheRollback(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil, logrus.New())
	assert.Nil(t, err)

	accountTrieCache := storagemgr.NewCacheWrapper(32, true)
	storageTrieCache := storagemgr.NewCacheWrapper(32, true)
	tc := NewPruneCache(createMockRepo(t), pStateStorage, accountTrieCache, storageTrieCache, logger)

	batch := pStateStorage.NewBatch()

	trieJournal1 := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type1")},
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v1"),
					"k2": makeLeafNode("v2"),
				},
				PruneSet: map[string]struct{}{},
			},
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type1")},
				DirtySet: map[string]types.Node{
					"k3": makeLeafNode("v3"),
				},
				PruneSet: map[string]struct {
				}{
					"k4": {},
				},
			},
		},
	}
	tc.Update(batch, 1, trieJournal1)

	trieJournal2 := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type1")},
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v11"),
					"k5": makeLeafNode("v55"),
				},
				PruneSet: map[string]struct{}{
					"k2": {},
				},
			},
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type1")},
				DirtySet:    map[string]types.Node{},
				PruneSet: map[string]struct {
				}{
					"k3": {},
				},
			},
		},
	}
	tc.Update(batch, 2, trieJournal2)

	batch.Commit()
	batch.Reset()

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
	trieJournal2 = &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type1")},
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v11"),
					"k5": makeLeafNode("v55"),
				},
				PruneSet: map[string]struct{}{
					"k2": {},
				},
			},
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type1")},
				DirtySet:    map[string]types.Node{},
				PruneSet: map[string]struct{}{
					"k3": {},
				},
			},
		},
	}
	tc.Update(batch, 2, trieJournal2)
	batch.Commit()

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

func TestPruningFlushByMaxBlockNum(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil, logrus.New())
	assert.Nil(t, err)

	accountTrieCache := storagemgr.NewCacheWrapper(32, true)
	storageTrieCache := storagemgr.NewCacheWrapper(32, true)
	tc := NewPruneCache(createMockRepo(t), pStateStorage, accountTrieCache, storageTrieCache, logger)

	batch := pStateStorage.NewBatch()

	for i := 0; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1; i++ {
		trieJournal1 := &types.StateDelta{
			Journal: []*types.TrieJournal{
				{
					RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
					RootNodeKey: &types.NodeKey{Version: uint64(i + 1), Path: []byte("path"), Type: []byte("type")},
					DirtySet: map[string]types.Node{
						"k" + strconv.Itoa(i+1): makeLeafNode("v" + strconv.Itoa(i+1)),
					},
					PruneSet: map[string]struct{}{},
				},
			},
		}
		tc.Update(batch, uint64(i+1), trieJournal1)
	}
	time.Sleep(100 * time.Millisecond)
	v, ok := tc.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1), []byte("k"+strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1)))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v"+strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1)), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	trieJournal2 := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), Path: []byte("path"), Type: []byte("type")},
				DirtySet: map[string]types.Node{
					"k" + strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum): makeLeafNode("v" + strconv.Itoa(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum)),
				},
			},
		},
	}
	tc.Update(batch, uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), trieJournal2)
	time.Sleep(2*checkFlushTimeInterval + 300*time.Millisecond)
	v, ok = tc.Get(1, []byte("k1"))
	require.False(t, ok)
	require.Nil(t, v)
	require.Equal(t, makeLeafNode("v1").Encode(), tc.ledgerStorage.Get([]byte("k1"))) // backend flushed
	require.Equal(t, tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(tc.states.diffs))

}

func TestPruningFlushByMaxBatchSize(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil, logrus.New())
	assert.Nil(t, err)

	accountTrieCache := storagemgr.NewCacheWrapper(32, true)
	storageTrieCache := storagemgr.NewCacheWrapper(32, true)
	pruneCache := NewPruneCache(createMockRepo(t), pStateStorage, accountTrieCache, storageTrieCache, logger)

	batch := pStateStorage.NewBatch()

	bigV := make([]byte, maxFlushBatchSizeThreshold)
	rand.Read(bigV)
	trieJournal := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 1, Path: []byte("path"), Type: []byte("type")},
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode(string(bigV)),
				},
			},
		},
	}
	pruneCache.Update(batch, uint64(1), trieJournal)

	for i := 1; i < pruneCache.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum; i++ {
		trieJournal := &types.StateDelta{
			Journal: []*types.TrieJournal{
				{
					RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
					RootNodeKey: &types.NodeKey{Version: uint64(i + 1), Path: []byte("path"), Type: []byte("type")},
					DirtySet: map[string]types.Node{
						"k1": makeLeafNode("v1"),
					},
				},
			},
		}
		pruneCache.Update(batch, uint64(i+1), trieJournal)
	}
	time.Sleep(time.Second)
	v, ok := pruneCache.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	v, ok = pruneCache.Get(uint64(pruneCache.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	require.Nil(t, pruneCache.ledgerStorage.Get([]byte("k1"))) // not flush

	trieJournal2 := &types.StateDelta{
		Journal: []*types.TrieJournal{
			{
				RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
				RootNodeKey: &types.NodeKey{Version: 2, Path: []byte("path"), Type: []byte("type")},
				DirtySet: map[string]types.Node{
					"k1": makeLeafNode("v11"),
				},
			},
		},
	}
	pruneCache.Update(batch, uint64(pruneCache.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), trieJournal2)
	time.Sleep(5*checkFlushTimeInterval + 300*time.Millisecond)
	v, ok = pruneCache.Get(uint64(pruneCache.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v11"), v)
	v1 := pruneCache.ledgerStorage.Get([]byte("k1"))
	require.True(t, len(v1) > 0)
	require.Equal(t, makeLeafNode(string(bigV)).Encode(), v1) // flushed
	require.Equal(t, pruneCache.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum, len(pruneCache.states.diffs))

}

func TestPruningFlushByMaxFlushTime(t *testing.T) {
	logger := log.NewWithModule("prune_test")
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil, logrus.New())
	assert.Nil(t, err)

	accountTrieCache := storagemgr.NewCacheWrapper(32, true)
	storageTrieCache := storagemgr.NewCacheWrapper(32, true)
	tc := NewPruneCache(createMockRepo(t), pStateStorage, accountTrieCache, storageTrieCache, logger)
	batch := pStateStorage.NewBatch()

	for i := 0; i < tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+maxFlushBlockNum-1; i++ {
		trieJournal := &types.StateDelta{
			Journal: []*types.TrieJournal{
				{
					RootHash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
					RootNodeKey: &types.NodeKey{Version: uint64(i + 1), Path: []byte("path"), Type: []byte("type")},
					DirtySet: map[string]types.Node{
						"k1": makeLeafNode("v1"),
					},
				},
			},
		}
		tc.Update(batch, uint64(i+1), trieJournal)
	}
	time.Sleep(100 * time.Millisecond)
	v, ok := tc.Get(uint64(1), []byte("k1"))
	require.True(t, ok)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	require.Nil(t, tc.ledgerStorage.Get([]byte("k1"))) // not flush

	time.Sleep(maxFlushTimeInterval + checkFlushTimeInterval + 100*time.Millisecond)
	v, ok = tc.Get(uint64(tc.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum+1), []byte("k1"))
	require.True(t, ok)
	require.Equal(t, makeLeafNode("v1"), v)
	v1 := tc.ledgerStorage.Get([]byte("k1"))
	require.True(t, len(v1) > 0)
	require.Equal(t, makeLeafNode("v1").Encode(), v1) // flushed
	require.Equal(t, 10, len(tc.states.diffs))

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
	// speed up unit test
	{
		checkFlushTimeInterval = 10 * time.Millisecond
		maxFlushBlockNum = 5
		maxFlushTimeInterval = 5 * time.Second
		maxFlushBatchSizeThreshold = 1 * 1024 * 1024 // 1MB
	}
	return r
}
