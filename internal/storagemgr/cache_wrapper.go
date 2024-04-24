package storagemgr

import (
	"github.com/coocood/freecache"
)

type CacheWrapper struct {
	cache *freecache.Cache

	metrics *CacheMetrics

	enableMetric bool
}

type CacheMetrics struct {
	CacheHitCounter  int
	CacheMissCounter int

	CacheSize uint64 // total bytes
}

func NewCacheWrapper(megabytesLimit int, enableMetric bool) *CacheWrapper {
	if megabytesLimit <= 0 {
		megabytesLimit = 128
	}

	return &CacheWrapper{
		cache:        freecache.NewCache(megabytesLimit * 1024 * 1024),
		metrics:      &CacheMetrics{},
		enableMetric: enableMetric,
	}
}

func (c *CacheWrapper) ResetCounterMetrics() {
	c.metrics.CacheMissCounter = 0
	c.metrics.CacheHitCounter = 0
}

func (c *CacheWrapper) ExportMetrics() *CacheMetrics {
	//var s fastcache.Stats
	//c.cache.UpdateStats(&s)
	//c.metrics.CacheSize = s.BytesSize
	return c.metrics
}

func (c *CacheWrapper) Get(k []byte) ([]byte, bool) {
	res, err := c.cache.Get(k)
	if err == freecache.ErrNotFound {
		c.metrics.CacheHitCounter++
	} else {
		c.metrics.CacheMissCounter++
	}
	return res, err != freecache.ErrNotFound
}

func (c *CacheWrapper) Set(k []byte, v []byte) {
	c.cache.Set(k, v, 0)
}

func (c *CacheWrapper) Del(k []byte) {
	c.cache.Del(k)
}

func (c *CacheWrapper) Reset() {
	c.cache.Clear()
	c.metrics = &CacheMetrics{}
}
