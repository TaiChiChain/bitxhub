package eth

import "github.com/prometheus/client_golang/prometheus"

var (
	queryTotalCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jsonrpc",
		Name:      "query_total_times",
		Help:      "the total query times for calculate query-per-second (qps)",
	})

	queryFailedCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jsonrpc",
		Name:      "query_failed_times",
		Help:      "the query failed times",
	})

	invokeReadOnlyDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jsonrpc",
		Name:      "invoke_read_only_latency",
		Help:      "The latency of invoking read-only rpc interface",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
	})

	invokeCallContractDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jsonrpc",
		Name:      "invoke_call_contract_latency",
		Help:      "The latency of invoking call contract rpc interface",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	invokeSendRawTxDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jsonrpc",
		Name:      "invoke_send_raw_tx_latency",
		Help:      "The latency of invoking send raw tx rpc interface",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
	})
)

func init() {
	prometheus.MustRegister(queryTotalCounter)
	prometheus.MustRegister(queryFailedCounter)
	prometheus.MustRegister(invokeReadOnlyDuration)
	prometheus.MustRegister(invokeCallContractDuration)
	prometheus.MustRegister(invokeSendRawTxDuration)
}
