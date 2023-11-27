package ledger

import "github.com/prometheus/client_golang/prometheus"

var (
	persistBlockDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "persist_block_duration_second",
		Help:      "The total latency of block persist",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	blockHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "block_height",
		Help:      "the latest block height",
	})

	flushDirtyWorldStateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "flush_dirty_world_state_duration",
		Help:      "The total latency of flush dirty world state into db",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	accountFlushSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "flush_dirty_account_size",
		Help:      "The size of flush dirty account into db",
	})

	stateFlushSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "flush_dirty_state_size",
		Help:      "The size of flush dirty state into db",
	})

	accountCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "account_cache_hit_counter_per_block",
		Help:      "The total number of account cache hit per block",
	})

	accountCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "account_cache_miss_counter_per_block",
		Help:      "The total number of account cache miss per block",
	})

	accountReadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "account_read_duration",
		Help:      "The total latency of read an account from db",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})

	stateReadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "state_read_duration",
		Help:      "The total latency of read a state from db",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})

	codeReadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "code_read_duration",
		Help:      "The total latency of read a contract code from db",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})
)

func init() {
	prometheus.MustRegister(persistBlockDuration)
	prometheus.MustRegister(blockHeightMetric)
	prometheus.MustRegister(flushDirtyWorldStateDuration)
	prometheus.MustRegister(accountCacheHitCounterPerBlock)
	prometheus.MustRegister(accountCacheMissCounterPerBlock)
	prometheus.MustRegister(accountReadDuration)
	prometheus.MustRegister(stateReadDuration)
	prometheus.MustRegister(accountFlushSize)
	prometheus.MustRegister(stateFlushSize)
	prometheus.MustRegister(codeReadDuration)
}
