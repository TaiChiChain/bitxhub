package adaptor

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type networkMessageMetrics struct {
	// monitor every proposal's header size
	headerSize prometheus.Summary

	// monitor every batch size
	recvBatchSize prometheus.Summary

	// monitor every request batch message size
	requestBatchSize prometheus.Summary

	// latency from request to response
	requestLatency prometheus.Histogram
}

const latencyAge = time.Minute
const latencyAgeBuckets = 1

var latencySecBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.2, 1.4,
	1.6, 1.8, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0, 5.5, 6.0, 6.5, 7.0, 7.5, 8.0, 8.5, 9.0, 9.5, 10.,
	12.5, 15., 17.5, 20., 25., 30., 60., 90., 120., 180., 300.}

var latencyObjectives = map[float64]float64{0.5: 0.05}

func newNetworkMessageMetrics() *networkMessageMetrics {
	m := &networkMessageMetrics{}
	m.recvBatchSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "dagbft",
			Name:       "batch_size",
			Help:       "Size in bytes of transaction batch",
			Objectives: map[float64]float64{0.5: 0.05},
			MaxAge:     1 * time.Minute,
			AgeBuckets: 1,
		})

	m.headerSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "dagbft",
			Name:       "header_size",
			Help:       "Size in bytes of proposed header",
			Objectives: map[float64]float64{0.5: 0.05},
			MaxAge:     1 * time.Minute,
			AgeBuckets: 1,
		})
	m.requestBatchSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "dagbft",
			Name:       "request_batch_size",
			Help:       "Size in bytes of requested batches",
			Objectives: map[float64]float64{0.5: 0.05},
			MaxAge:     1 * time.Minute,
			AgeBuckets: 1,
		})

	m.requestLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "request_latency",
			Help:      "The latency of request",
			Buckets:   latencySecBuckets,
		})

	prometheus.MustRegister(m.recvBatchSize)
	prometheus.MustRegister(m.headerSize)
	prometheus.MustRegister(m.requestBatchSize)
	prometheus.MustRegister(m.requestLatency)
	return m
}

type blockChainMetrics struct {
	syncChainCounter      prometheus.Counter
	discardedTransactions *prometheus.CounterVec
	batchVerifyLatency    prometheus.Summary
	batchTxsCounter       prometheus.Counter
	batchCounter          prometheus.Counter
}

func newBlockChainMetrics() *blockChainMetrics {
	bm := &blockChainMetrics{}
	bm.syncChainCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "sync_chain_count",
			Help:      "The number of sync chain",
		},
	)

	bm.discardedTransactions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "discarded_transactions_num",
			Help:      "The number of discarded transactions",
		},
		[]string{"reason"},
	)

	bm.batchVerifyLatency = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "axiom_ledger",
			Subsystem:  "dagbft",
			Name:       "batch_verify_latency",
			Help:       "The latency of batch verify",
			Objectives: latencyObjectives,
			MaxAge:     latencyAge,
			AgeBuckets: latencyAgeBuckets,
		},
	)

	bm.batchTxsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "batch_txs_count",
			Help:      "The number of batch's transactions",
		},
	)

	bm.batchCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "axiom_ledger",
			Subsystem: "dagbft",
			Name:      "batch_count",
			Help:      "The number of batch in consensus output",
		},
	)

	prometheus.MustRegister(bm.syncChainCounter)
	prometheus.MustRegister(bm.discardedTransactions)
	prometheus.MustRegister(bm.batchVerifyLatency)
	prometheus.MustRegister(bm.batchTxsCounter)
	prometheus.MustRegister(bm.batchCounter)
	return bm
}
