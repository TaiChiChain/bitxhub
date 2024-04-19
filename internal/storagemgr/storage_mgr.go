package storagemgr

import (
	"fmt"
	"runtime"
	"sync"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
	"github.com/prometheus/common/model"

	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/storage/kv/leveldb"
	"github.com/axiomesh/axiom-kit/storage/kv/pebble"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	BlockChain = "blockchain"
	Ledger     = "ledger"
	Snapshot   = "snapshot"
	Blockfile  = "blockfile"
	Consensus  = "consensus"
	Epoch      = "epoch"
	TxPool     = "txpool"
	Sync       = "sync"
)

var globalStorageMgr = &storageMgr{
	storageBuilderMap: make(map[string]func(p string, metricsPrefixName string) (kv.Storage, error)),
	storages:          make(map[string]kv.Storage),
	lock:              new(sync.Mutex),
}

func init() {
	memoryBuilder := func(p string, metricsPrefixName string) (kv.Storage, error) {
		return kv.NewMemory(), nil
	}

	// only for test
	globalStorageMgr.storageBuilderMap[repo.KVStorageTypeLeveldb] = memoryBuilder
	globalStorageMgr.storageBuilderMap[repo.KVStorageTypePebble] = memoryBuilder
	globalStorageMgr.storageBuilderMap[""] = memoryBuilder
}

type storageMgr struct {
	storageBuilderMap map[string]func(p string, metricsPrefixName string) (kv.Storage, error)
	storages          map[string]kv.Storage
	defaultKVType     string
	lock              *sync.Mutex
}

var defaultPebbleOptions = &pebbledb.Options{
	// MemTableStopWritesThreshold is max number of the existent MemTables(including the frozen one).
	// This manner is the same with leveldb, including a frozen memory table and another live one.
	MemTableStopWritesThreshold: 2,

	// The default compaction concurrency(1 thread)
	MaxConcurrentCompactions: func() int { return runtime.NumCPU() },

	// Per-level options. Options for at least one level must be specified. The
	// options for the last level are used for all subsequent levels.
	// This option is the same with Ethereum.
	Levels: []pebbledb.LevelOptions{
		{TargetFileSize: 2 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 2 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 4 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 4 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 8 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 8 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		{TargetFileSize: 16 * 1024 * 1024, BlockSize: 32 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
	},
}

func (m *storageMgr) open(typ string, p string, metricsPrefixName string) (kv.Storage, error) {
	builder, ok := m.storageBuilderMap[typ]
	if !ok {
		return nil, fmt.Errorf("unknow kv type %s, expect leveldb or pebble", typ)
	}
	return builder(p, metricsPrefixName)
}

func Initialize(defaultKVType string, defaultKvCacheSize int, sync bool, enableMetrics bool) error {
	globalStorageMgr.storageBuilderMap[repo.KVStorageTypeLeveldb] = func(p string, _ string) (kv.Storage, error) {
		return leveldb.New(p, nil)
	}
	globalStorageMgr.storageBuilderMap[repo.KVStorageTypePebble] = func(p string, metricsPrefixName string) (kv.Storage, error) {
		defaultPebbleOptions.Cache = pebbledb.NewCache(int64(defaultKvCacheSize * 1024 * 1024))
		defaultPebbleOptions.MemTableSize = defaultKvCacheSize * 1024 * 1024 / 4 // The size of single memory table
		namespace := "axiom_ledger"
		subsystem := "ledger"
		var metricOpts []pebble.MetricsOption
		if enableMetrics && model.IsValidMetricName(model.LabelValue(metricsPrefixName)) {
			metricOpts = append(metricOpts,
				pebble.WithDiskSizeGauge(namespace, subsystem, metricsPrefixName),
				pebble.WithDiskWriteThroughput(namespace, subsystem, metricsPrefixName),
				pebble.WithWalWriteThroughput(namespace, subsystem, metricsPrefixName),
				pebble.WithEffectiveWriteThroughput(namespace, subsystem, metricsPrefixName))
		}
		return pebble.New(p, defaultPebbleOptions, &pebbledb.WriteOptions{Sync: sync}, loggers.Logger(loggers.Storage), metricOpts...)
	}
	_, ok := globalStorageMgr.storageBuilderMap[defaultKVType]
	if !ok {
		return fmt.Errorf("unknow kv type %s, expect leveldb or pebble", defaultKVType)
	}
	globalStorageMgr.defaultKVType = defaultKVType
	return nil
}

func Open(p string) (kv.Storage, error) {
	return OpenSpecifyType(globalStorageMgr.defaultKVType, p, "")
}

func OpenWithMetrics(p string, uniqueMetricsPrefixName string) (kv.Storage, error) {
	if uniqueMetricsPrefixName != "" && !model.IsValidMetricName(model.LabelValue(uniqueMetricsPrefixName)) {
		return nil, fmt.Errorf("%q is not a valid metric name", uniqueMetricsPrefixName)
	}
	return OpenSpecifyType(globalStorageMgr.defaultKVType, p, uniqueMetricsPrefixName)
}

func OpenSpecifyType(typ string, p string, metricName string) (kv.Storage, error) {
	globalStorageMgr.lock.Lock()
	defer globalStorageMgr.lock.Unlock()
	s, ok := globalStorageMgr.storages[p]
	if !ok {
		var err error
		s, err = globalStorageMgr.open(typ, p, metricName)
		if err != nil {
			return nil, err
		}
		globalStorageMgr.storages[p] = s
	}
	return s, nil
}
