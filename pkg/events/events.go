package events

import (
	"github.com/axiomesh/axiom-kit/types"
)

type ExecutedEvent struct {
	Block         *types.Block
	TxPointerList []*TxPointer
}

type TxPointer struct {
	Hash    *types.Hash
	Account string
	Nonce   uint64
}
