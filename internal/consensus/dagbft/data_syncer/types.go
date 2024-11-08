package data_syncer

import (
	"container/heap"
	"fmt"

	"github.com/axiomesh/axiom-ledger/internal/components/axm_heap"
	"github.com/axiomesh/axiom-ledger/internal/components/status"
	consensusTypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/utils/concurrency"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
)

type event int

const (
	newAttestation event = iota
	newTxSet
	stateUpdate
	executed
	syncEpoch
	finishSyncEpoch
	finishStateUpdate
)

var EventTypeM = map[event]string{
	newAttestation:    "newAttestation",
	newTxSet:          "newTxSet",
	stateUpdate:       "stateUpdate",
	executed:          "executed",
	syncEpoch:         "syncEpoch",
	finishSyncEpoch:   "finishSyncEpoch",
	finishStateUpdate: "finishStateUpdate",
}

const (
	Normal status.StatusType = iota
	InEpochSync
	InSync
)

const (
	cacheAbnormal status.StatusType = iota
	cacheFull
)

type bindNode struct {
	wsAddr string
	p2pID  string
	nodeId uint64
}
type localEvent struct {
	EventType event
	Event     any
}

type ProofCache struct {
	quorumCheckpointM map[uint64]*consensusTypes.DagbftQuorumCheckpoint
	heightIndex       *axm_heap.NumberHeap
	maxSize           int
	status            *status.StatusMgr
}

func newProofCache(maxSize int) *ProofCache {
	ac := &ProofCache{
		quorumCheckpointM: make(map[uint64]*consensusTypes.DagbftQuorumCheckpoint),
		heightIndex:       new(axm_heap.NumberHeap),
		maxSize:           maxSize,
		status:            status.NewStatusMgr(),
	}
	*ac.heightIndex = make([]uint64, 0)
	heap.Init(ac.heightIndex)

	return ac
}

func (ac *ProofCache) pushProof(quorumCkpt *consensusTypes.DagbftQuorumCheckpoint, metrics *metrics) error {
	defer ac.checkStatus()
	if !ac.isNormal() {
		return fmt.Errorf("cache status is abnormal")
	}
	ac.quorumCheckpointM[quorumCkpt.Height()] = quorumCkpt
	ac.heightIndex.PushItem(quorumCkpt.Height())
	metrics.attestationCacheNum.Inc()
	return nil
}

type stateUpdateEvent struct {
	targetCheckpoint *types.QuorumCheckpoint
	EpochChanges     []*types.QuorumCheckpoint
}

func newAsyncPool(size int, lg logrus.FieldLogger) (*concurrency.AsyncPool, error) {
	pool, err := ants.NewPool(size, ants.WithLogger(&concurrency.AsyncPoolLogger{Printer: func(s string, a ...any) { lg.Panicf(s, a...) }}))
	if err != nil {
		return nil, err
	}
	ap := &concurrency.AsyncPool{
		Pool: pool,
	}

	return ap, nil
}
