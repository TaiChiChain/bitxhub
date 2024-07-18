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
)

func init() {
	prometheus.MustRegister(blockStmStepDuration)
}
