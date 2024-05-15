package archive

import (
	"sync"
	"time"

	"github.com/axiomesh/axiom-bft/common/consensus"
	rbfttypes "github.com/axiomesh/axiom-bft/types"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/status"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
)

const (
	syncStateRestart    timer.TimeoutEvent = "syncStateRestart"
	fetchMissingTxsResp timer.TimeoutEvent = "fetchMissingTxsResp"
	syncStateResp       timer.TimeoutEvent = "syncStateResp"
)

const (
	waitResponseTimeout = 1 * time.Second
	maxCacheSize        = 10000
	maxRetryCount       = 3
)

const (
	Normal status.StatusType = iota
	NeedSyncState
	InSyncState
	NeedFetchMissingTxs
	InCommit
	StateTransferring
	InEpochSyncing
)

var statusTypes = map[status.StatusType]string{
	Normal:            "Normal",
	NeedSyncState:     "NeedSyncState",
	InSyncState:       "InSyncState",
	InCommit:          "InCommit",
	StateTransferring: "StateTransferring",
	InEpochSyncing:    "InEpochSyncing",
}

const (
	eventType_commitToExecutor = iota
	eventType_syncBlock
	eventType_stateUpdated
	eventType_executed
	eventType_epochSync
	eventType_consensusMessage
)

var eventTypes = map[int]string{
	eventType_commitToExecutor: "commitToExecutor",
	eventType_syncBlock:        "syncBlock",
	eventType_stateUpdated:     "stateUpdated",
	eventType_executed:         "executed",
	eventType_epochSync:        "epochSync",
	eventType_consensusMessage: "consensusMessage",
}

type archiveEvent struct {
	EventType int
	Event     any
}

type chainConfig struct {
	epochInfo *types.EpochInfo
	view      uint64
	H         uint64       // Low watermark block number.
	lock      sync.RWMutex // mutex to set value
}

type readyExecute[T any, Constraint types.TXConstraint[T]] struct {
	txs       []*T
	localList []bool
	height    uint64
	timestamp int64
	proposer  uint64
}

type wrapFetchMissingRequest struct {
	request    *consensus.FetchMissingRequest
	proposer   uint64
	retryCount int
}

type stateUpdateTarget struct {
	// target height and digest
	metaState *rbfttypes.MetaState

	// signed checkpoints that prove the above target
	checkpointSet []*consensus.SignedCheckpoint

	// path of epoch changes from epoch-change-proof
	epochChanges []*consensus.EpochChange
}
