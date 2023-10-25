package storagemgr

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	pebble2 "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
)

const (
	BlockChain = "blockchain"
	Ledger     = "ledger"
	Blockfile  = "blockfile"
	Consensus  = "consensus"
	Epoch      = "epoch"
)

var globalStorageMgr = &storageMgr{
	storageBuilderMap: make(map[string]func(p string) (storage.Storage, error)),
	storages:          make(map[string]storage.Storage),
	lock:              new(sync.Mutex),
}

type storageMgr struct {
	storageBuilderMap map[string]func(p string) (storage.Storage, error)
	storages          map[string]storage.Storage
	defaultKVType     string
	lock              *sync.Mutex
}

var defaultPebbleOptions = &pebble2.Options{
	// MemTableStopWritesThreshold is max number of the existent MemTables(including the frozen one).
	// This manner is the same with leveldb, including a frozen memory table and another live one.
	MemTableStopWritesThreshold: 2,

	// The default compaction concurrency(1 thread)
	MaxConcurrentCompactions: func() int { return runtime.NumCPU() },

	// Per-level options. Options for at least one level must be specified. The
	// options for the last level are used for all subsequent levels.
	// This option is the same with Ethereum.
	Levels: []pebble2.LevelOptions{
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
	},
}

func (m *storageMgr) Open(typ string, p string) (storage.Storage, error) {
	builder, ok := m.storageBuilderMap[typ]
	if !ok {
		return nil, fmt.Errorf("unknow kv type %s, expect leveldb or pebble", typ)
	}
	return builder(p)
}

func Initialize(defaultKVType string, defaultKvCacheSize int) error {
	globalStorageMgr.storageBuilderMap[repo.KVStorageTypeLeveldb] = func(p string) (storage.Storage, error) {
		return leveldb.New(p, nil)
	}
	globalStorageMgr.storageBuilderMap[repo.KVStorageTypePebble] = func(p string) (storage.Storage, error) {
		defaultPebbleOptions.Cache = pebble2.NewCache(int64(defaultKvCacheSize * 1024 * 1024))
		defaultPebbleOptions.MemTableSize = defaultKvCacheSize * 1024 * 1024 / 4 // The size of single memory table
		return pebble.New(p, defaultPebbleOptions)
	}
	_, ok := globalStorageMgr.storageBuilderMap[defaultKVType]
	if !ok {
		return fmt.Errorf("unknow kv type %s, expect leveldb or pebble", defaultKVType)
	}
	globalStorageMgr.defaultKVType = defaultKVType
	return nil
}

func Open(p string) (storage.Storage, error) {
	return OpenSpecifyType(globalStorageMgr.defaultKVType, p)
}

func OpenSpecifyType(typ string, p string) (storage.Storage, error) {
	globalStorageMgr.lock.Lock()
	defer globalStorageMgr.lock.Unlock()
	s, ok := globalStorageMgr.storages[p]
	if !ok {
		var err error
		s, err = globalStorageMgr.Open(typ, p)
		if err != nil {
			return nil, err
		}
		globalStorageMgr.storages[p] = s
	}
	return s, nil
}
