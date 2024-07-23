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

	getTransactionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "get_transaction_duration",
		Help:      "The total latency of get a transaction from db",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})

	getOrCreateAccountDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "get_or_create_account_duration",
		Help:      "The total latency of get or create account",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})

	getBalanceDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "get_balance_duration",
		Help:      "The total latency of get balance",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})

	setBalanceDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "set_balance_duration",
		Help:      "The total latency of set balance",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 10),
	})

	cacheHit = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "cache_hit",
		Help:      "The total number of cache hit",
	})

	snapHit = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "snap_hit",
		Help:      "The total number of snap hit",
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

	prometheus.MustRegister(getOrCreateAccountDuration)
	prometheus.MustRegister(getBalanceDuration)
	prometheus.MustRegister(setBalanceDuration)
	prometheus.MustRegister(cacheHit)
	prometheus.MustRegister(snapHit)
}
