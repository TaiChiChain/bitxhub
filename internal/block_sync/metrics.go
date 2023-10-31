package block_sync

import "github.com/prometheus/client_golang/prometheus"

var (
	blockSyncDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "axiom_ledger",
			Subsystem: "block_sync",
			Name:      "block_sync_duration_seconds",
			Help:      "The total latency of block sync",
		},
		[]string{"sync_count"},
	)

	requesterNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "block_sync",
			Name:      "requester_number",
			Help:      "The total number of requester",
		},
	)

	recvBlockNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "block_sync",
			Name:      "recv_block_number",
			Help:      "The recv Blcok number of every chunk",
		},
		[]string{"chunk_size"},
	)

	invalidBlockNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "axiom_ledger",
			Subsystem: "block_sync",
			Name:      "invalid_Block_number",
			Help:      "The total number of invalid block",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(blockSyncDuration)
	prometheus.MustRegister(requesterNumber)
	prometheus.MustRegister(invalidBlockNumber)
	prometheus.MustRegister(recvBlockNumber)
}
