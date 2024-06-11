package sync

import "github.com/prometheus/client_golang/prometheus"

var (
	blockSyncDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "axiom_ledger",
			Subsystem: "sync",
			Name:      "block_sync_duration_seconds",
			Help:      "The total latency of commitData sync",
		},
		[]string{"sync_count"},
	)

	validateBlockDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "axiom_ledger",
			Subsystem: "sync",
			Name:      "validate_block_duration_seconds",
			Help:      "The total latency of commitData validation",
		},
	)

	requesterNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "sync",
			Name:      "requester_number",
			Help:      "The total number of requester",
		},
	)

	recvBlockNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "sync",
			Name:      "recv_block_number",
			Help:      "The recv Blcok number of every chunk",
		},
		[]string{"chunk_size"},
	)

	invalidBlockNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "sync",
			Name:      "invalid_Block_number",
			Help:      "The total number of invalid commitData",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(blockSyncDuration)
	prometheus.MustRegister(validateBlockDuration)
	prometheus.MustRegister(requesterNumber)
	prometheus.MustRegister(invalidBlockNumber)
	prometheus.MustRegister(recvBlockNumber)
}
