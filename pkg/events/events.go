package events

import (
	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
)

type ExecutedEvent struct {
	Block                  *types.Block
	TxPointerList          []*TxPointer
	StateUpdatedCheckpoint *consensus.Checkpoint
}

type TxPointer struct {
	Hash    *types.Hash
	Account string
	Nonce   uint64
}
