package events

import (
	"github.com/axiomesh/axiom-kit/types"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
)

type ExecutedEvent struct {
	Block                  *types.Block
	TxPointerList          []*TxPointer
	StateUpdatedCheckpoint *consensustypes.Checkpoint
	CommitSequence         uint64
}

type TxPointer struct {
	Hash    *types.Hash
	Account string
	Nonce   uint64
}
