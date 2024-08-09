package blockstm

import "github.com/prometheus/client_golang/prometheus"

var (
	blockStmStepDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "block_stm_step_duration_second",
		Help:      "The total latency of blockstm step",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	blockStmSettleDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "block_stm_settle_duration_second",
		Help:      "The total latency of blockstm settle",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	mvHashmapReadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "mv_hashmap_read_duration_second",
		Help:      "The total latency of mv_hashmap read",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	mvHashmapReadNullDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "mv_hashmap_read_null_duration_second",
		Help:      "The total latency of mv_hashmap read null",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	mvHashmapReadWithOutLockDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "mv_hashmap_read_without_lock_duration_second",
		Help:      "The total latency of mv_hashmap read without lock",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	mvHashmapWriteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "mv_hashmap_write_duration_second",
		Help:      "The total latency of mv_hashmap write",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	blockStmExecuteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "block_stm_execute_duration_second",
		Help:      "The total latency of blockstm execute",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	blockStmValidateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "block_stm_validate_duration_second",
		Help:      "The total latency of blockstm validate",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	blockStmFlushDataDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "block_stm_flush_data_duration_second",
		Help:      "The total latency of blockstm flush data",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
)

func init() {
	prometheus.MustRegister(blockStmStepDuration)
	prometheus.MustRegister(blockStmSettleDuration)
	prometheus.MustRegister(mvHashmapReadDuration)
	prometheus.MustRegister(mvHashmapWriteDuration)
	prometheus.MustRegister(mvHashmapReadWithOutLockDuration)
	prometheus.MustRegister(mvHashmapReadNullDuration)
	prometheus.MustRegister(blockStmExecuteDuration)
	prometheus.MustRegister(blockStmValidateDuration)
	prometheus.MustRegister(blockStmFlushDataDuration)
}
