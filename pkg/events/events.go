package events

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
)

type ExecutedEvent struct {
	Block                  *types.Block
	TxPointerList          []*TxPointer
	StateUpdatedCheckpoint *common.Checkpoint
}

type TxPointer struct {
	Hash    *types.Hash
	Account string
	Nonce   uint64
}
