package common

import (
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

const (
	LocalTxEvent = iota
	RemoteTxEvent
)

var (
	ErrorPreCheck       = errors.New("precheck failed")
	ErrorAddTxPool      = errors.New("add txpool failed")
	ErrorConsensusStart = errors.New("consensus not start yet")
)

// UncheckedTxEvent represents misc event sent by local modules
type UncheckedTxEvent struct {
	EventType int
	Event     any
}

type TxWithResp struct {
	Tx      *types.Transaction
	CheckCh chan *TxResp
	PoolCh  chan *TxResp
}

type TxResp struct {
	Status   bool
	ErrorMsg string
}

type CommitEvent struct {
	Block                  *types.Block
	StateUpdatedCheckpoint *consensus.Checkpoint
}

func StoreEpochState(epochStore storage.Storage, key string, value []byte) error {
	epochStore.Put([]byte("epoch."+key), value)
	return nil
}

func ReadEpochState(epochStore storage.Storage, key string) ([]byte, error) {
	b := epochStore.Get([]byte("epoch." + key))
	if b == nil {
		return nil, errors.New("not found")
	}
	return b, nil
}
