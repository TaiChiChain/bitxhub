package solo

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
)

const (
	singleTx    = 1
	maxChanSize = 1024
	ErrPoolFull = "txpool is full"
)

// consensusEvent is a type meant to clearly convey that the return type or parameter to a function will be supplied to/from an events.Manager
type consensusEvent any

// chainState is a type for reportState
type chainState struct {
	Height     uint64
	BlockHash  *types.Hash
	TxHashList []*types.Hash
}

// getLowWatermarkReq is a type for syncer request GetLowWatermark
type getLowWatermarkReq struct {
	Resp chan uint64
}

type genBatchReq struct {
	typ int
}

type batchTimerManager struct {
	timer.Timer
	lastBatchTime           int64
	minTimeoutBatchTime     float64
	minNoTxTimeoutBatchTime float64
}
