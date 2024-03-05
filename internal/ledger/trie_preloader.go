package ledger

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage"
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

	// subPreloaderPool use sync.pool to reduce subpreloader malloc
	subPreloaderPool = sync.Pool{
		New: func() any {
			return &subPreloader{}
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

// triePreloader is an trie preloader, which is preload trie node to avoid
// huge trie loading when commit state
type triePreloader struct {
	logger    logrus.FieldLogger
	db        storage.Storage // load trie node would be cached in this db
	trieCache jmt.TrieCache
	rootHash  common.Hash

	loaders map[string]*subPreloader
}

func newTriePreloader(logger logrus.FieldLogger, db storage.Storage, trieCache jmt.TrieCache, rootHash common.Hash) *triePreloader {
	return &triePreloader{
		logger:    logger,
		db:        db,
		trieCache: trieCache,
		rootHash:  rootHash,
		loaders:   make(map[string]*subPreloader),
	}
}

func (tp *triePreloader) close() {
	for _, loader := range tp.loaders {
		loader.close()
	}
	tp.loaders = nil
}

// preload schedules sub preloader to preload tries
func (tp *triePreloader) preload(rootHash common.Hash, keys [][]byte) {
	if rootHash == (common.Hash{}) {
		rootHash = tp.rootHash
	}
	trieKey := tp.trieKey(rootHash)
	if tp.loaders == nil {
		tp.loaders = make(map[string]*subPreloader)
	}
	loader := tp.loaders[trieKey]
	if loader == nil {
		loader = newSubPreloader(tp.logger, tp.db, tp.trieCache, rootHash)
		tp.loaders[trieKey] = loader
	}
	loader.schedule(keys)
}

func (tp *triePreloader) trieKey(rootHash common.Hash) string {
	return string(rootHash.Bytes())
}

// subPreloader is a trie preload goroutinue for get a single trie.
// The subPreloader would process all requests from the same trie.
type subPreloader struct {
	logger    logrus.FieldLogger
	db        storage.Storage
	trieCache jmt.TrieCache
	rootHash  common.Hash
	trie      *jmt.JMT

	preloadKeys [][]byte
	lock        sync.Mutex

	wake chan struct{}
	stop chan struct{}
	term chan struct{}

	cached map[string]struct{}
}

func newSubPreloader(logger logrus.FieldLogger, db storage.Storage, trieCache jmt.TrieCache, rootHash common.Hash) *subPreloader {
	sp := subPreloaderPool.Get().(*subPreloader)
	sp.logger = logger
	sp.db = db
	sp.rootHash = rootHash
	sp.trieCache = trieCache

	sp.wake = make(chan struct{}, 1)
	sp.stop = make(chan struct{})
	sp.term = make(chan struct{})
	sp.cached = make(map[string]struct{})

	go sp.loop()
	return sp
}

func (sp *subPreloader) schedule(keys [][]byte) {
	// append the preload keys to the current queue
	sp.lock.Lock()
	sp.preloadKeys = append(sp.preloadKeys, keys...)
	sp.lock.Unlock()

	// notify the preloader to load
	select {
	case sp.wake <- struct{}{}:
	default:
	}
}

func (sp *subPreloader) close() {
	defer func() {
		sp.reset()
		subPreloaderPool.Put(sp)
	}()

	select {
	case <-sp.stop:
	default:
		close(sp.stop)
	}
	<-sp.term
}

func (sp *subPreloader) reset() {
	sp.logger = nil
	sp.db = nil
	sp.rootHash = common.Hash{}
	sp.trie = nil
	sp.preloadKeys = sp.preloadKeys[:0]

	sp.wake = make(chan struct{}, 1)
	sp.stop = make(chan struct{})
	sp.term = make(chan struct{})
	sp.cached = make(map[string]struct{})
}

func (sp *subPreloader) loop() {
	defer close(sp.term)

	if sp.trie == nil {
		trie, err := jmt.New(sp.rootHash, sp.db, sp.trieCache, sp.logger)
		if err != nil {
			sp.logger.Errorf("Trie preloader failed to new jmt trie, root hash: %s", sp.rootHash)
			return
		}
		sp.trie = trie
	}

	for {
		select {
		case <-sp.wake:
			sp.lock.Lock()
			preloadKeys := sp.preloadKeys
			sp.preloadKeys = nil
			sp.lock.Unlock()

			for i, key := range preloadKeys {
				select {
				case <-sp.stop:
					sp.lock.Lock()
					sp.preloadKeys = append(sp.preloadKeys, preloadKeys[i:]...)
					sp.lock.Unlock()
					return

				default:
					if _, ok := sp.cached[string(key)]; !ok {
						// haven't preload
						_, err := sp.trie.Get(key)
						if err != nil {
							sp.logger.Errorf("Load trie node error, key: %s", key)
							continue
						}
						sp.cached[string(key)] = struct{}{}
						triePreloadMissCountPerBlock++
					} else {
						triePreloadHitCountPerBlock++
					}
				}
			}

		case <-sp.stop:
			return
		}
	}
}
