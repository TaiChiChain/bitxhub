package txcache

import (
	"github.com/prometheus/client_golang/prometheus"
)

type txCacheMetrics struct {
	// monitor all incoming txs
	incomingTxsCounter prometheus.Counter

	// monitor all pending txs, which have been batched into set but
	// has not been delivered to consensus core.
	pendingTxs prometheus.Gauge

	// monitor latency of receiving tx from api
	txAPILatency *prometheus.SummaryVec
}

func newTxCacheMetrics() *txCacheMetrics {
	m := &txCacheMetrics{}
	m.incomingTxsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "txCache",
			Name:      "incoming_txs",
			Help:      "incoming txs counter of txset",
		})

	m.pendingTxs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "txCache",
			Name:      "pending_txs",
			Help:      "pending txs in txset",
		})

	m.txAPILatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "txCache",
			Name:       "api_latency",
			Help:       "latency of receiving transactions from api",
			Objectives: map[float64]float64{0.5: 0.05},
		}, []string{"type"},
	)

	prometheus.MustRegister(m.incomingTxsCounter)
	prometheus.MustRegister(m.pendingTxs)
	prometheus.MustRegister(m.txAPILatency)

	return m
}
