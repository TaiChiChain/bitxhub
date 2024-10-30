package data_syncer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	ErrorType              = "invalid type"
	ErrorHandleNewTxSet    = "handle new tx set failed"
	ErrorHandleNewBlock    = "handle new block failed"
	ErrorHandleStateUpdate = "handle state update failed"
)

type metrics struct {
	// monitor every transaction propagation
	txCounter prometheus.Counter

	// monitor every new attestation size
	attestationSize prometheus.Summary

	attestationCacheNum prometheus.Counter

	// monitor every state update
	stateUpdatesNum *prometheus.CounterVec

	epochChangeNum prometheus.Counter

	errCounter *prometheus.CounterVec
}

const latencyAge = time.Minute
const latencyAgeBuckets = 1

var latencySecBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.2, 1.4,
	1.6, 1.8, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0, 5.5, 6.0, 6.5, 7.0, 7.5, 8.0, 8.5, 9.0, 9.5, 10.,
	12.5, 15., 17.5, 20., 25., 30., 60., 90., 120., 180., 300.}

var latencyObjectives = map[float64]float64{0.5: 0.05}

func newMetrics() *metrics {
	m := &metrics{}
	m.attestationSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "dagbft",
			Name:       "attestation_size",
			Help:       "Size in bytes of every new attestation",
			Objectives: map[float64]float64{0.5: 0.05},
			MaxAge:     1 * time.Minute,
			AgeBuckets: 1,
		})

	m.txCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "tx_count",
			Help:      "Number of transactions",
		},
	)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "error_count",
			Help:      "Number of errors",
		}, []string{"reason"})

	m.stateUpdatesNum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "state_update_count",
			Help:      "Number of state updates",
		}, []string{"reason"})

	m.attestationCacheNum = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "attestation_cache_count",
			Help:      "Number of attestation cache",
		},
	)
	m.epochChangeNum = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "epoch_change_count",
			Help:      "Number of epoch change",
		},
	)
	prometheus.MustRegister(m.txCounter)
	prometheus.MustRegister(m.attestationSize)
	prometheus.MustRegister(m.stateUpdatesNum)
	prometheus.MustRegister(m.errCounter)
	prometheus.MustRegister(m.attestationCacheNum)
	prometheus.MustRegister(m.epochChangeNum)
	return m
}
