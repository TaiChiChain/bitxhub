package ledger

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage/kv"
)

var (
	triePreloadHitCountPerBlock  int
	triePreloadMissCountPerBlock int

	triePreloadHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "storage",
		Name:      "trie_preload_hit_counter_per_block",
		Help:      "The total number of trie node preload hit per block",
	})

	triePreloadMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "storage",
		Name:      "trie_preload_miss_counter_per_block",
		Help:      "The total number of trie node preload miss per block",
	})

	// preloaderPool use sync.pool to reduce preloader malloc
	preloaderPool = sync.Pool{
		New: func() any {
			return &triePreloader{}
		},
	}
)

func init() {
	prometheus.MustRegister(triePreloadHitCounterPerBlock)
	prometheus.MustRegister(triePreloadMissCounterPerBlock)
}

func ExportTriePreloaderMetrics() {
	triePreloadHitCounterPerBlock.Set(float64(triePreloadHitCountPerBlock))
	triePreloadMissCounterPerBlock.Set(float64(triePreloadMissCountPerBlock))
}

func ResetTriePreloaderMetrics() {
	triePreloadHitCountPerBlock = 0
	triePreloadMissCountPerBlock = 0
}

// triePreloaderManager manage lifecycle of all trie preloaders
type triePreloaderManager struct {
	logger     logrus.FieldLogger
	db         kv.Storage
	pruneCache jmt.PruneCache
	trieCache  jmt.TrieCache

	loaders map[string]*triePreloader

	wg sync.WaitGroup
}

func newTriePreloaderManager(logger logrus.FieldLogger, db kv.Storage, trieCache jmt.TrieCache, pruneCache jmt.PruneCache) *triePreloaderManager {
	return &triePreloaderManager{
		logger:     logger,
		db:         db,
		pruneCache: pruneCache,
		trieCache:  trieCache,
		loaders:    make(map[string]*triePreloader),
		wg:         sync.WaitGroup{},
	}
}

func (tp *triePreloaderManager) close() {
	for _, loader := range tp.loaders {
		loader.close()
	}
	tp.loaders = nil
}

func (tp *triePreloaderManager) wait() {
	tp.wg.Wait()
	tp.wg = sync.WaitGroup{}
}

// preload loads keys from trie whose root is identified by rootHash
func (tp *triePreloaderManager) preload(rootHash common.Hash, keys [][]byte) {
	// empty trie, don't need to preload
	if rootHash == (common.Hash{}) {
		return
	}

	trieKey := string(rootHash.Bytes())
	if tp.loaders == nil {
		tp.loaders = make(map[string]*triePreloader)
	}
	loader := tp.loaders[trieKey]
	if loader == nil {
		loader = newPreloader(tp.logger, tp.db, tp.trieCache, tp.pruneCache, rootHash, &tp.wg)
		tp.loaders[trieKey] = loader
		tp.wg.Add(1)
		go loader.loop()
	}

	// append the preload keys to the current queue
	loader.lock.Lock()
	loader.preloadKeys = append(loader.preloadKeys, keys...)
	loader.lock.Unlock()

	// notify the preloader to load
	select {
	case loader.wake <- struct{}{}:
	default:
	}
}

// triePreloader is a trie preload goroutinue for get a single trie.
// The triePreloader would process all requests from the same trie.
type triePreloader struct {
	logger logrus.FieldLogger

	db         kv.Storage
	pruneCache jmt.PruneCache
	trieCache  jmt.TrieCache

	rootHash    common.Hash
	trie        *jmt.JMT
	preloadKeys [][]byte
	cached      map[string]struct{}

	lock sync.Mutex
	wg   *sync.WaitGroup
	wake chan struct{}
}

func newPreloader(logger logrus.FieldLogger, db kv.Storage, trieCache jmt.TrieCache, pruneCache jmt.PruneCache, rootHash common.Hash, wg *sync.WaitGroup) *triePreloader {
	sp := preloaderPool.Get().(*triePreloader)
	sp.logger = logger
	sp.db = db
	sp.rootHash = rootHash
	sp.pruneCache = pruneCache
	sp.trieCache = trieCache
	sp.wg = wg

	sp.wake = make(chan struct{}, 1)
	sp.cached = make(map[string]struct{})

	return sp
}

func (loader *triePreloader) close() {
	defer func() {
		loader.reset()
		preloaderPool.Put(loader)
	}()
}

func (loader *triePreloader) reset() {
	loader.logger = nil
	loader.db = nil
	loader.rootHash = common.Hash{}
	loader.trie = nil
	loader.preloadKeys = loader.preloadKeys[:0]
	loader.wake = make(chan struct{}, 1)
	loader.cached = make(map[string]struct{})
	loader.wg = nil
}

func (loader *triePreloader) loop() {
	defer loader.wg.Done()

	if loader.trie == nil {
		trie, err := jmt.New(loader.rootHash, loader.db, loader.trieCache, loader.pruneCache, loader.logger)
		if err != nil {
			loader.logger.Errorf("Trie preloader failed to new jmt trie, root hash: %s", loader.rootHash)
			return
		}
		loader.trie = trie
	}

	for {
		select {
		case <-loader.wake:
			loader.lock.Lock()
			preloadKeys := loader.preloadKeys
			loader.preloadKeys = nil
			loader.lock.Unlock()

			for _, key := range preloadKeys {
				if _, ok := loader.cached[string(key)]; !ok {
					// haven't preload
					_, err := loader.trie.Get(key)
					if err != nil {
						loader.logger.Errorf("Load trie node error, key: %s", key)
						continue
					}
					loader.cached[string(key)] = struct{}{}
					triePreloadMissCountPerBlock++
				} else {
					triePreloadHitCountPerBlock++
				}
			}

			return
		}
	}
}
