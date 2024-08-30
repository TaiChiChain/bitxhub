package executor

import "github.com/prometheus/client_golang/prometheus"

var (
	applyTxsDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "apply_transactions_duration_seconds",
		Help:      "The total latency of transactions apply",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	})
	commitBlockDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "commit_block_duration_seconds",
		Help:      "The total latency of block commit",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	cs2ExecutorDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "cs_to_executor_duration_seconds",
		Help:      "The total latency of consensus to executor",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	executeBlockDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "execute_block_duration_second",
		Help:      "The total latency of block execute",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	calcMerkleDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "calc_merkle_duration_seconds",
		Help:      "The total latency of merkle calc",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	txCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "tx_counter",
		Help:      "the total number of transactions",
	})
	proposedBlockCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "proposed_block_counter",
		Help:      "the total number of node proposed blocks",
	})
	perBlockTxCounter = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "per_block_tx_counter",
		Help:      "the total number of transactions per block",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"epoch"})
)

func init() {
	prometheus.MustRegister(applyTxsDuration)
	prometheus.MustRegister(calcMerkleDuration)
	prometheus.MustRegister(executeBlockDuration)
	prometheus.MustRegister(commitBlockDuration)
	prometheus.MustRegister(cs2ExecutorDuration)
	prometheus.MustRegister(txCounter)
	prometheus.MustRegister(proposedBlockCounter)
	prometheus.MustRegister(perBlockTxCounter)
}
