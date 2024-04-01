package snapshot

import "github.com/prometheus/client_golang/prometheus"

var (
	snapshotAccountCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "snapshot",
		Name:      "snapshot_account_cache_miss_counter_per_block",
		Help:      "The total number of snapshot account cache miss per block",
	})

	snapshotAccountCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "snapshot",
		Name:      "snapshot_account_cache_hit_counter_per_block",
		Help:      "The total number of snapshot account cache hit per block",
	})

	snapshotAccountCacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "snapshot",
		Name:      "snapshot_account_cache_size",
		Help:      "The total size of snapshot account cache (MB)",
	})
)

func init() {
	prometheus.MustRegister(snapshotAccountCacheMissCounterPerBlock)
	prometheus.MustRegister(snapshotAccountCacheHitCounterPerBlock)
	prometheus.MustRegister(snapshotAccountCacheSize)
}
