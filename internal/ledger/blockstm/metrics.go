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

	mvHashmapWriteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "ledger",
		Name:      "mv_hashmap_write_duration_second",
		Help:      "The total latency of mv_hashmap write",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
)

func init() {
	prometheus.MustRegister(blockStmStepDuration)
	prometheus.MustRegister(blockStmSettleDuration)
	prometheus.MustRegister(mvHashmapReadDuration)
	prometheus.MustRegister(mvHashmapWriteDuration)
}
