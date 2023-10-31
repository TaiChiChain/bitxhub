package solo

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	batchInterval = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "solo",
			Name:       "batch_interval_duration",
			Help:       "interval duration of batch",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type"},
	)

	minBatchIntervalDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "solo",
			Name:      "min_batch_interval_duration",
			Help:      "min timeout interval duration of batch",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(batchInterval)
	prometheus.MustRegister(minBatchIntervalDuration)
}
