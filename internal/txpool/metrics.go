package txpool

import "github.com/prometheus/client_golang/prometheus"

var (
	poolTxNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "txpool",
			Name:      "tx_counter",
			Help:      "the total number of transactions",
		},
	)
	readyTxNum = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "txpool",
		Name:      "ready_tx_counter",
		Help:      "the total number of transactions which ready to generate batch",
	})
	processEventDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "txpool",
			Name:      "process_event_duration_seconds",
			Help:      "the duration of process event",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
		},
		[]string{"event"},
	)
	rejectTxNum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "txpool",
			Name:      "reject_tx_counter",
			Help:      "the total number of rejected transactions",
		},
		[]string{"reason"},
	)
	removeTxNum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "txpool",
			Name:      "remove_tx_counter",
			Help:      "the total number of transactions which removed from txpool",
		},
		[]string{"reason"},
	)
)

func init() {
	prometheus.MustRegister(processEventDuration)
	prometheus.MustRegister(poolTxNum)
	prometheus.MustRegister(readyTxNum)
	prometheus.MustRegister(rejectTxNum)
	prometheus.MustRegister(removeTxNum)
}
