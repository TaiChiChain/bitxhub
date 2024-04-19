package storagemgr

import (
	"github.com/VictoriaMetrics/fastcache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/axiomesh/axiom-kit/storage/kv"
)

var (
	kvCacheHitCountPerBlock  int
	kvCacheMissCountPerBlock int

	kvCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "storage",
		Name:      "kv_cache_hit_counter_per_block",
		Help:      "The total number of kv cache hit per block",
	})

	kvCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "storage",
		Name:      "kv_cache_miss_counter_per_block",
		Help:      "The total number of kv cache miss per block",
	})
)

func init() {
	prometheus.MustRegister(kvCacheHitCounterPerBlock)
	prometheus.MustRegister(kvCacheMissCounterPerBlock)
}

func ExportCachedStorageMetrics() {
	kvCacheHitCounterPerBlock.Set(float64(kvCacheHitCountPerBlock))
	kvCacheMissCounterPerBlock.Set(float64(kvCacheMissCountPerBlock))
}

func ResetCachedStorageMetrics() {
	kvCacheHitCountPerBlock = 0
	kvCacheMissCountPerBlock = 0
}

type CachedStorage struct {
	kv.Storage
	cache *fastcache.Cache
}

func NewCachedStorage(s kv.Storage, megabytesLimit int) kv.Storage {
	if megabytesLimit <= 0 {
		megabytesLimit = 128
	}
	return &CachedStorage{
		Storage: s,
		cache:   fastcache.New(megabytesLimit * 1024 * 1024),
	}
}

func (c *CachedStorage) Get(key []byte) []byte {
	value, ok := c.cache.HasGet(nil, key)
	if ok {
		kvCacheHitCountPerBlock++
		return value
	}
	v := c.Storage.Get(key)
	kvCacheMissCountPerBlock++
	if v != nil {
		c.cache.Set(key, v)
	}
	return v
}

func (c *CachedStorage) Has(key []byte) bool {
	has := c.cache.Has(key)
	if has {
		kvCacheHitCountPerBlock++
		return true
	}
	kvCacheMissCountPerBlock++
	return c.Storage.Has(key)
}

func (c *CachedStorage) Put(key, value []byte) {
	if len(value) == 0 {
		value = nil
	}
	c.Storage.Put(key, value)
	c.cache.Set(key, value)
}

func (c *CachedStorage) Delete(key []byte) {
	c.cache.Del(key)
	c.Storage.Delete(key)
}

func (c *CachedStorage) Close() error {
	c.cache.Reset()
	return c.Storage.Close()
}

func (c *CachedStorage) NewBatch() kv.Batch {
	return &BatchWrapper{
		Batch:      c.Storage.NewBatch(),
		cache:      c.cache,
		finalState: make(map[string][]byte),
	}
}

type BatchWrapper struct {
	kv.Batch
	cache      *fastcache.Cache
	finalState map[string][]byte
}

func (w *BatchWrapper) Put(key, value []byte) {
	if len(value) == 0 {
		w.finalState[string(key)] = nil
		w.Batch.Delete(key)
	} else {
		w.finalState[string(key)] = value
		w.Batch.Put(key, value)
	}
}

func (w *BatchWrapper) Delete(key []byte) {
	w.finalState[string(key)] = nil
	w.Batch.Delete(key)
}

func (w *BatchWrapper) Commit() {
	w.Batch.Commit()
	for k, v := range w.finalState {
		if v == nil {
			w.cache.Del([]byte(k))
		} else {
			w.cache.Set([]byte(k), v)
		}
	}
}

func (w *BatchWrapper) Size() int {
	return w.Batch.Size()
}

func (w *BatchWrapper) Reset() {
	w.Batch.Reset()
	w.finalState = make(map[string][]byte)
}
