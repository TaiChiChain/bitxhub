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
	evmExecuteBlockDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "evm_execute_block_duration_second",
		Help:      "The total latency of block execute",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})
	evmExecuteEachDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "evm_execute_each_duration_second",
		Help:      "The tx latency of evm execute",
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
	calcBlockSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "executor",
		Name:      "calc_block_size",
		Help:      "The size of current block calc",
		Buckets:   prometheus.ExponentialBuckets(1024, 2, 12),
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
)

func init() {
	prometheus.MustRegister(applyTxsDuration)
	prometheus.MustRegister(calcMerkleDuration)
	prometheus.MustRegister(calcBlockSize)
	prometheus.MustRegister(evmExecuteBlockDuration)
	prometheus.MustRegister(evmExecuteEachDuration)
	prometheus.MustRegister(executeBlockDuration)
	prometheus.MustRegister(txCounter)
	prometheus.MustRegister(proposedBlockCounter)
}
