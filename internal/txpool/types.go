package txpool

import (
	"math/big"
	"time"

	common_pool "github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
)

// nolint
const (
	DefaultPoolSize               = 50000
	DefaultToleranceNonceGap      = 1000
	DefaultToleranceTime          = 5 * time.Minute
	DefaultToleranceRemoveTime    = 15 * time.Minute
	DefaultCleanEmptyAccountTime  = 10 * time.Minute
	DefaultRotateTxLocalsInterval = 1 * time.Hour

	maxChanSize = 1024
)

type GetAccountNonceFunc func(address string) uint64

type GetAccountBalanceFunc func(address string) *big.Int

type txPointer struct {
	account string
	nonce   uint64
}

type internalTransaction[T any, Constraint types.TXConstraint[T]] struct {
	rawTx       *T
	local       bool
	lifeTime    int64 // track the local txs' broadcast time
	arrivedTime int64 // track the local txs' arrived txpool time
}

type txPoolEvent any

// =====================addTxsEvent=============================
const (
	localTxEvent = iota
	remoteTxsEvent
	reBroadcastTxsEvent
	missingTxsEvent
	localRecordTxEvent
)

var addTxsEventToStr = map[int]string{
	localTxEvent:        "localTxEvent",
	remoteTxsEvent:      "remoteTxsEvent",
	reBroadcastTxsEvent: "reBroadcastTxsEvent",
	missingTxsEvent:     "missingTxsEvent",
	localRecordTxEvent:  "localRecordTxEvent",
}

// LocalEvent represents event sent by local modules
type addTxsEvent struct {
	EventType int
	Event     any
}

type reqLocalTx[T any, Constraint types.TXConstraint[T]] struct {
	tx    *T
	errCh chan error
}

type reqLocalRecordTx[T any, Constraint types.TXConstraint[T]] struct {
	txs []*T
	ch  chan int
}

type reqRemoteTxs[T any, Constraint types.TXConstraint[T]] struct {
	txs []*T
}

type reqMissingTxs[T any, Constraint types.TXConstraint[T]] struct {
	batchHash string
	txs       map[uint64]*T
	errCh     chan error
}

// ================================================================

// =====================removeTxsEvent=============================
const (
	timeoutTxsEvent = iota
	highNonceTxsEvent
	committedTxsEvent
	batchedTxsEvent
	invalidTxsEvent
)

var removeTxsEventToStr = map[int]string{
	timeoutTxsEvent:   "timeoutTxsEvent",
	highNonceTxsEvent: "highNonceTxsEvent",
	committedTxsEvent: "committedTxsEvent",
	batchedTxsEvent:   "batchedTxsEvent",
	invalidTxsEvent:   "invalidTxsEvent",
}

type removeTxsEvent struct {
	EventType int
	Event     any
}

type reqHighNonceTxs struct {
	account   string
	highNonce uint64
}

type reqRemoveCommittedTxs struct {
	txPointerList []*common_pool.WrapperTxPointer
}

type reqRemoveBatchedTxs struct {
	batchHashList []string
}

type reqRemoveInvalidTxs[T any, Constraint types.TXConstraint[T]] struct {
	removeTxs map[string][]*internalTransaction[T, Constraint]
}

// ========================================================================

// ========================batchEvent===================================
var batchEventToStr = map[int]string{
	common_pool.GenBatchTimeoutEvent:     "GenBatchTimeoutEvent",
	common_pool.GenBatchNoTxTimeoutEvent: "GenBatchNoTxTimeoutEvent",
	common_pool.GenBatchFirstEvent:       "GenBatchFirstEvent",
	common_pool.GenBatchSizeEvent:        "GenBatchSizeEvent",
	common_pool.ReConstructBatchEvent:    "ReConstructBatchEvent",
	common_pool.GetTxsForGenBatchEvent:   "GetTxsForGenBatchEvent",
}

type batchEvent struct {
	EventType int
	Event     any
}

type reqGenBatch[T any, Constraint types.TXConstraint[T]] struct {
	respCh chan *respGenBatch[T, Constraint]
}

type respGenBatch[T any, Constraint types.TXConstraint[T]] struct {
	resp *common_pool.RequestHashBatch[T, Constraint]
	err  error
}

type reqReConstructBatch[T any, Constraint types.TXConstraint[T]] struct {
	oldBatch *common_pool.RequestHashBatch[T, Constraint]
	respCh   chan *respReConstructBatch
}

type respReConstructBatch struct {
	deDuplicateTxHashes []string
	err                 error
}

type reqGetTxsForGenBatch[T any, Constraint types.TXConstraint[T]] struct {
	batchHash           string
	timestamp           int64
	hashList            []string
	deDuplicateTxHashes []string

	respCh chan *respGetTxsForGenBatch[T, Constraint]
}

type respGetTxsForGenBatch[T any, Constraint types.TXConstraint[T]] struct {
	txs            []*T
	localList      []bool
	missingTxsHash map[uint64]string
	err            error
}

// ========================================================================

// =======================consensusEvent=============================
const (
	SendMissingTxsEvent = iota
	FilterReBroadcastTxsEvent
	RestoreOneBatchEvent
	RestoreAllBatchedEvent
)

var consensusEventToStr = map[int]string{
	SendMissingTxsEvent:       "SendMissingTxsEvent",
	FilterReBroadcastTxsEvent: "FilterReBroadcastTxsEvent",
	RestoreOneBatchEvent:      "RestoreOneBatchEvent",
	RestoreAllBatchedEvent:    "RestoreAllBatchedEvent",
}

type consensusEvent struct {
	EventType int
	Event     any
}

type reqSendMissingTxs[T any, Constraint types.TXConstraint[T]] struct {
	batchHash       string
	missingHashList map[uint64]string

	respCh chan *respSendMissingTxs[T, Constraint]
}

type respSendMissingTxs[T any, Constraint types.TXConstraint[T]] struct {
	resp map[uint64]*T
	err  error
}

type reqFilterReBroadcastTxs[T any, Constraint types.TXConstraint[T]] struct {
	timeout bool
	respCh  chan []*T
}

type reqRestoreOneBatch struct {
	batchHash string
	errCh     chan error
}

// =========================poolInfoEvent===============================
const (
	reqTxEvent = iota
	reqNonceEvent
	reqPendingTxCountEvent
	reqPoolMetaEvent
	reqAccountMetaEvent
	reqChainInfoEvent
	updateChainInfoEvent
)

var poolInfoEventToStr = map[int]string{
	reqTxEvent:             "reqTxEvent",
	reqNonceEvent:          "reqNonceEvent",
	reqPendingTxCountEvent: "reqPendingTxCountEvent",
	reqPoolMetaEvent:       "reqPoolMetaEvent",
	reqAccountMetaEvent:    "reqAccountMetaEvent",
	reqChainInfoEvent:      "reqChainInfoEvent",
	updateChainInfoEvent:   "updateChainInfoEvent",
}

// poolInfoEvent represents poolInfo event sent by local api modules
type poolInfoEvent struct {
	EventType int
	Event     any
}

type reqTxMsg[T any, Constraint types.TXConstraint[T]] struct {
	hash string
	ch   chan *T
}

type reqNonceMsg struct {
	account string
	ch      chan uint64
}

type reqPendingTxCountMsg struct {
	ch chan uint64
}

type reqAccountPoolMetaMsg[T any, Constraint types.TXConstraint[T]] struct {
	account string
	full    bool
	ch      chan *common_pool.AccountMeta[T, Constraint]
}

type reqPoolMetaMsg[T any, Constraint types.TXConstraint[T]] struct {
	full bool
	ch   chan *common_pool.Meta[T, Constraint]
}

type reqChainInfoMsg struct {
	ch chan *common_pool.ChainInfo
}

type updateChainInfoMsg struct {
	chainInfo *common_pool.ChainInfo
}

// =========================localEvent===============================
const (
	gcAccountEvent = iota
	rotateTxLocalsEvent
)

var localEventToStr = map[int]string{
	gcAccountEvent:      "gcAccountEvent",
	rotateTxLocalsEvent: "rotateTxLocals",
}

type localEvent struct {
	EventType int
	Event     any
}
