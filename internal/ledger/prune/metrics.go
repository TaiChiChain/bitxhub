package prune

import "github.com/prometheus/client_golang/prometheus"

var (
	pruneCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "prune",
		Name:      "prune_cache_miss_counter_per_block",
		Help:      "The total number of prune cache miss per block",
	})

	pruneCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "prune",
		Name:      "prune_cache_hit_counter_per_block",
		Help:      "The total number of prune cache hit per block",
	})
)

func init() {
	prometheus.MustRegister(pruneCacheMissCounterPerBlock)
	prometheus.MustRegister(pruneCacheHitCounterPerBlock)
}
