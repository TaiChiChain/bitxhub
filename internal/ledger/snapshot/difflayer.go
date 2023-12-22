package snapshot

//
//import (
//	"bytes"
//	"math/big"
//	"sync"
//
//	"github.com/ethereum/go-ethereum/common"
//
//	"github.com/axiomesh/axiom-kit/types"
//)
//
//type diffLayer struct {
//	parent Layer // Parent snapshot modified by this one, never nil
//	origin Layer // bottom disk layer
//
//	stateRoot common.Hash // Block StateRoot hash associated with the base snapshot
//	available bool
//
//	destructs map[string]struct{}
//	accounts  map[string][]byte
//	storage   map[string]map[string][]byte
//
//	lock sync.RWMutex
//}
//
//// Root returns stateRoot hash for which this snapshot was made.
//func (dl *diffLayer) Root() common.Hash {
//	return dl.stateRoot
//}
//
//// Parent always returns nil as there's no layer below the disk.
//func (dl *diffLayer) Parent() Layer {
//	return dl.parent
//}
//
//func (dl *diffLayer) Available() bool {
//	dl.lock.RLock()
//	defer dl.lock.RUnlock()
//
//	return dl.available
//}
//
//// parameter 'storage' means <contract address --> <storage key --> storage value>>
//func (dl *diffLayer) New(parent Layer, stateRoot common.Hash, destructs map[string]struct{}, accounts map[string][]byte, storage map[string]map[string][]byte) Layer {
//	l := &diffLayer{
//		parent:    parent,
//		stateRoot: stateRoot,
//		destructs: destructs,
//		accounts:  accounts,
//		storage:   storage,
//	}
//	return l
//}
//
//func (dl *diffLayer) Account(addr *types.Address) (*types.InnerAccount, error) {
//	dl.lock.RLock()
//	defer dl.lock.RUnlock()
//
//	if dl.available {
//		return nil, ErrSnapshotUnavailable
//	}
//
//	// find target account in current layer
//	if blob, ok := dl.accounts[addr.String()]; ok {
//		innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
//		if err := innerAccount.Unmarshal(blob); err != nil {
//			panic(err)
//		}
//		return innerAccount, nil
//	}
//
//	if _, ok := dl.destructs[addr.String()]; ok {
//		return nil, nil
//	}
//
//	// find target account in parent layer recursively
//	if dl.parent != nil {
//		return dl.parent.Account(addr)
//	}
//
//	// find target account in disk layer
//	return dl.origin.Account(addr)
//}
//
//func (dl *diffLayer) Storage(addr *types.Address, key []byte) ([]byte, error) {
//	dl.lock.RLock()
//	defer dl.lock.RUnlock()
//
//	if dl.available {
//		return nil, ErrSnapshotUnavailable
//	}
//
//	// find target storage slot in current layer
//	if slots, ok := dl.storage[addr.String()]; ok {
//		if v, ok := slots[string(key)]; ok {
//			return v, nil
//		}
//	}
//
//	if _, ok := dl.destructs[addr.String()]; ok {
//		return nil, nil
//	}
//
//	// find target storage slot in parent layer recursively
//	if dl.parent != nil {
//		return dl.parent.Storage(addr, key)
//	}
//
//	// find target storage slot in disk layer
//	return dl.origin.Storage(addr, key)
//}
//
//func (dl *diffLayer) Journal(buffer *bytes.Buffer) (common.Hash, error) {
//	return common.Hash{}, nil
//}
//
//func (dl *diffLayer) Update(stateRoot common.Hash, destructs map[string]struct{}, accounts map[string]*types.InnerAccount, storage map[string]map[string][]byte) {
//
//}
