package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	SendTx2ConsensusCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "send_tx_to_consensus_counter",
		Help:      "The number of transactions send to consensus",
	})

	Consensus2ExecuteBlockTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "cs_to_execute_block_time",
		Help:      "The latency of consensus to executor block",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"consensus"})

	BatchVerifyTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "batch_verify_time",
		Help:      "The latency of batch verify",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"consensus"})
	WaitEpochTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "wait_epoch_time",
		Help:      "The latency of wait epoch",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"consensus"})
	ExecutedBlockCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "executed_block_counter",
		Help:      "The number of executed blocks",
	}, []string{"consensus"})
	Consensus2ExecutorBlockCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "consensus_to_executor_block_counter",
		Help:      "The number of consensus committed to executor blocks",
	}, []string{"consensus"})
	BatchCommitLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "batch_commit_latency",
		Help:      "The latency of batch commit",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"consensus", "type"})
	// ProcessConsensusMessageDuration monitor the time from network request to consensus return result.
	ProcessConsensusMessageDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "process_consensus_message_duration",
		Help:      "The latency of process consensus message",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"consensus", "event"})
)

func init() {
	prometheus.MustRegister(SendTx2ConsensusCounter)
	prometheus.MustRegister(Consensus2ExecuteBlockTime)
	prometheus.MustRegister(BatchVerifyTime)
	prometheus.MustRegister(WaitEpochTime)
	prometheus.MustRegister(ExecutedBlockCounter)
	prometheus.MustRegister(Consensus2ExecutorBlockCounter)
	prometheus.MustRegister(BatchCommitLatency)
	prometheus.MustRegister(ProcessConsensusMessageDuration)
}
