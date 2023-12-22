package snapshot

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/types"
)

var (
	ErrSnapshotUnavailable = errors.New("snapshot unavailable")
)

type Layer interface {
	// Root returns the stateRoot hash for which this snapshot was made.
	Root() common.Hash

	// Account directly retrieves the account associated with a particular hash in
	// the snapshot slim data format.
	Account(addr *types.Address) (*types.InnerAccount, error)

	// Storage directly retrieves the storage data associated with a particular hash,
	// within a particular account.
	Storage(addr *types.Address, key []byte) ([]byte, error)

	// Parent returns the subsequent layer of a snapshot, or nil if the base was
	// reached.
	//
	// Note, the method is an internal helper to avoid type switching between the
	// disk and diff layers. There is no locking involved.
	Parent() Layer

	// Available returns whether this layer is available.
	Available() bool

	Update(stateRoot common.Hash, destructs map[string]struct{}, accounts map[string]*types.InnerAccount, storage map[string]map[string][]byte)

	// Clear reset cache
	Clear()
}
