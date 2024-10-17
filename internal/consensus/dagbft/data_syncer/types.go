package data_syncer

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/heap"
)

type event int

const (
	newBlock event = iota
	newTxSet
)

type localEvent struct {
	EventType event
	Event     any
}

type blockCache struct {
	blockM      map[uint64]*types.Block
	heightIndex *heap.NumberHeap
}
