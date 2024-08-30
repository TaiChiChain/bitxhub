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

	accountTrieCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "account_trie_cache_miss_counter_per_block",
		Help:      "The total number of account trie cache miss per block",
	})

	accountTrieCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "account_trie_cache_hit_counter_per_block",
		Help:      "The total number of account trie cache hit per block",
	})

	accountTrieCacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "account_trie_cache_size",
		Help:      "The total size of account trie cache (MB)",
	})

	storageTrieCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "storage_trie_cache_miss_counter_per_block",
		Help:      "The total number of storage trie cache miss per block",
	})

	storageTrieCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "storage_trie_cache_hit_counter_per_block",
		Help:      "The total number of storage trie cache hit per block",
	})

	storageTrieCacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "storage_trie_cache_size",
		Help:      "The total size of storage trie cache (MB)",
	})

	getTransactionCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "ledger",
			Name:      "get_transaction_counter",
			Help:      "the total times of query GetTransaction",
		},
	)

	calcBlockSize = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "calc_block_size",
		Help:      "block size",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	})

	getTransactionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "get_transaction_duration",
		Help:      "The total latency of get a transaction from db",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})
)

func init() {
	prometheus.MustRegister(persistBlockDuration)
	prometheus.MustRegister(blockHeightMetric)
	prometheus.MustRegister(flushDirtyWorldStateDuration)
	prometheus.MustRegister(accountReadDuration)
	prometheus.MustRegister(stateReadDuration)
	prometheus.MustRegister(codeReadDuration)
	prometheus.MustRegister(accountTrieCacheMissCounterPerBlock)
	prometheus.MustRegister(accountTrieCacheHitCounterPerBlock)
	prometheus.MustRegister(accountTrieCacheSize)
	prometheus.MustRegister(storageTrieCacheMissCounterPerBlock)
	prometheus.MustRegister(storageTrieCacheHitCounterPerBlock)
	prometheus.MustRegister(storageTrieCacheSize)
	prometheus.MustRegister(getTransactionCounter)
	prometheus.MustRegister(getTransactionDuration)
	prometheus.MustRegister(calcBlockSize)
}
