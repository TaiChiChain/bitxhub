package blockstm

import "github.com/prometheus/client_golang/prometheus"

var (
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
	prometheus.MustRegister(blockStmSettleDuration)
	//prometheus.MustRegister(mvHashmapReadDuration)
	//prometheus.MustRegister(mvHashmapWriteDuration)
	prometheus.MustRegister(blockStmExecuteDuration)
	prometheus.MustRegister(blockStmValidateDuration)
	prometheus.MustRegister(blockStmFlushDataDuration)
}
