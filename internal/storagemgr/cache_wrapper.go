package storagemgr

import (
	"github.com/dgraph-io/ristretto"
)

type CacheWrapper struct {
	cache *ristretto.Cache

	metrics *CacheMetrics

	enableMetric bool
}

type CacheMetrics struct {
	CacheHitCounter  int
	CacheMissCounter int

	CacheSize uint64 // total bytes
}

var maxCost uint64

func NewCacheWrapper(megabytesLimit int, enableMetric bool) *CacheWrapper {
	if megabytesLimit <= 0 {
		megabytesLimit = 128
	}

	maxCost = uint64(megabytesLimit * (1 << 20))
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e8,            // number of keys to track frequency of (100M).
		MaxCost:     int64(maxCost), // maximum cost of cache (MB).
		BufferItems: 64,             // default config, no need to change
		Metrics:     true,
	})
	if err != nil {
		panic(err)
	}

	return &CacheWrapper{
		cache:        cache,
		metrics:      &CacheMetrics{},
		enableMetric: enableMetric,
	}
}

func (c *CacheWrapper) ResetCounterMetrics() {
	c.metrics.CacheMissCounter = 0
	c.metrics.CacheHitCounter = 0
}

func (c *CacheWrapper) ExportMetrics() *CacheMetrics {
	added := c.cache.Metrics.CostAdded()
	if added > maxCost {
		added = maxCost
	}
	c.metrics.CacheSize = added
	return c.metrics
}

func (c *CacheWrapper) Get(k []byte) ([]byte, bool) {
	res, ok := c.cache.Get(k)
	if ok {
		c.metrics.CacheHitCounter++
		return res.([]byte), ok
	}
	c.metrics.CacheMissCounter++
	return nil, ok
}

func (c *CacheWrapper) Set(k []byte, v []byte) {
	c.cache.Set(k, v, int64(len(k)+len(v)))
}

func (c *CacheWrapper) Del(k []byte) {
	c.cache.Del(k)
}

func (c *CacheWrapper) Reset() {
	c.cache.Clear()
	c.metrics = &CacheMetrics{}
}
