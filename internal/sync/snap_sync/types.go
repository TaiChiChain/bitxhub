package snap_sync

import (
	"github.com/axiomesh/axiom-kit/types"
	axm_heap "github.com/axiomesh/axiom-ledger/internal/components/heap"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
)

const concurrencyLimit = 100

type epochStateCache struct {
	quorumCheckpoints map[uint64]types.QuorumCheckpoint
	epochIndex        *axm_heap.NumberHeap
}

func newEpochStateCache() *epochStateCache {
	return &epochStateCache{
		quorumCheckpoints: make(map[uint64]types.QuorumCheckpoint),
		epochIndex:        new(axm_heap.NumberHeap),
	}
}

func (e *epochStateCache) getTopEpoch() uint64 {
	return e.epochIndex.PeekItem()
}

func (e *epochStateCache) popQuorumCheckpoint() containers.Pair[uint64, types.QuorumCheckpoint] {
	epoch := e.epochIndex.Pop().(uint64)
	value := e.quorumCheckpoints[epoch]
	delete(e.quorumCheckpoints, epoch)
	return containers.Pack2(epoch, value)
}

func (e *epochStateCache) pushQuorumCheckpoint(epoch uint64, value types.QuorumCheckpoint) {
	e.quorumCheckpoints[epoch] = value
	e.epochIndex.Push(epoch)
}
