package storagemgr

import (
	"testing"
	"time"

	"golang.org/x/exp/rand"
)

func TestCacheWrapperSpaceAmplify(t *testing.T) {
	// size := atomic.Uint64{}
	// cache := NewCacheWrapper(1024, true)
	//
	// wg := sync.WaitGroup{}
	// wg.Add(1)
	// go func() {
	//	for {
	//		time.Sleep(1 * time.Second)
	//		metrics := cache.ExportMetrics()
	//		fmt.Printf("real cache size = %v (KB)\n", metrics.CacheSize/1024)
	//		fmt.Printf("count cache size = %v (KB)\n", size.Load()/1024)
	//	}
	// }()
	//
	// wg.Add(1)
	// go func() {
	//	keys := make([][]byte, 0)
	//	threshold := 10000000
	//	for {
	//		k, v := getRandomHexKV(16, 16)
	//		if _, ok := cache.Get(k); !ok {
	//			cache.Set(k, v)
	//			keys = append(keys, k)
	//			size.Add(uint64(len(k) + len(v)))
	//		}
	//		if len(keys) > threshold {
	//			fmt.Printf("stop set cache\n")
	//			break
	//		}
	//	}
	//	for i := 0; i < threshold; i++ {
	//		cache.Del(keys[i])
	//	}
	// }()
	//
	// wg.Wait()
}

func getRandomHexKV(lk, lv int) (k []byte, v []byte) {
	rand.Seed(uint64(time.Now().UnixNano()))
	k = make([]byte, lk)
	v = make([]byte, lv)
	for i := 0; i < lk; i++ {
		k[i] = byte(rand.Intn(16))
	}
	for i := 0; i < lv; i++ {
		v[i] = byte(rand.Intn(16))
	}
	return k, v
}
