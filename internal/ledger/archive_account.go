package ledger

/*
#cgo LDFLAGS: /home/bcds/krc/forestore/crates/history_store/target/release/libhistory_store.a -lm -lstdc++
#include "/home/bcds/krc/forestore/crates/history_store/src/c_ffi/history.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"unsafe"
)

// #cgo LDFLAGS: /Users/zhangqirui/workplace/forestore/crates/history_store/target/release/libhistory_store.a -ldl -lm -lc++ -lc++abi
// #include "/Users/zhangqirui/workplace/forestore/crates/history_store/src/c_ffi/history.h"
// #include <stdlib.h>
type ArchiveAccount struct {
	archiveStateDB *C.EvmArchiveStateLedger

	addr *types.Address

	account *types.InnerAccount
}

func NewArchiveAccount(stateDBPtr *C.EvmArchiveStateLedger, address *types.Address, acc *types.InnerAccount) *ArchiveAccount {
	return &ArchiveAccount{
		archiveStateDB: stateDBPtr,
		addr:           address,
		account:        acc,
	}
}

func (o *ArchiveAccount) String() string {
	return fmt.Sprintf("{account: %v, code length: %v}", o.account, len(o.Code()))
}

// GetState Get state from local cache, if not found, then get it from DB
func (o *ArchiveAccount) GetState(key []byte) (bool, []byte) {
	var length C.uintptr_t
	ptr := C.get_storage_at_version(o.archiveStateDB, ConvertToCH256(o.account.StorageRoot),
		ConvertToCContractKey(o.addr, key), &length)
	val := C.GoBytes(unsafe.Pointer(ptr), C.int(length))
	return int(length) != 0, val
}

func (o *ArchiveAccount) GetCommittedState(key []byte) []byte {
	_, val := o.GetState(key)
	return val
}

// Code return the contract code
func (o *ArchiveAccount) Code() []byte {
	var length C.uintptr_t
	cCodeKey := ConvertToCCodeKey(o.addr, o.account.CodeHash)
	ptr := C.get_code(o.archiveStateDB, cCodeKey, &length)
	goSlice := C.GoBytes(unsafe.Pointer(ptr), C.int(length))
	return goSlice
}

func (o *ArchiveAccount) CodeHash() []byte {
	return o.account.CodeHash
}

// GetNonce Get the nonce from user account
func (o *ArchiveAccount) GetNonce() uint64 {
	return o.account.Nonce
}

// GetBalance Get the balance from the account
func (o *ArchiveAccount) GetBalance() *big.Int {
	return o.account.Balance
}

func (o *ArchiveAccount) GetAddress() *types.Address {
	return o.addr
}

// SetNonce Set the nonce which indicates the contract number
func (o *ArchiveAccount) SetNonce(nonce uint64) {
	panic("not support SetNonce")
}

// SetState Set account state
func (o *ArchiveAccount) SetState(k []byte, value []byte) {
	panic("not support SetState")

}

func (o *ArchiveAccount) SetBit256State(k []byte, value common.Hash) {
	panic("not support SetBit256State")

}

// SetCodeAndHash Set the contract code and hash
func (o *ArchiveAccount) SetCodeAndHash(code []byte) {
	panic("not support SetCodeAndHash")

}

// SetBalance Set the balance to the account
func (o *ArchiveAccount) SetBalance(balance *big.Int) {
	panic("not support SetBalance")

}

func (o *ArchiveAccount) SubBalance(amount *big.Int) {
	panic("not support SubBalance")

}

func (o *ArchiveAccount) AddBalance(amount *big.Int) {
	panic("not support AddBalance")

}

// Finalise moves all dirty states into the pending states.
// Return all dirty state keys
func (o *ArchiveAccount) Finalise() [][]byte {
	panic("not support Finalise")
}

func (o *ArchiveAccount) SetSelfDestructed(selfDestructed bool) {
	panic("not support SetSelfDestructed")
}

func (o *ArchiveAccount) IsEmpty() bool {
	return o.GetBalance().Sign() == 0 && o.GetNonce() == 0 && o.Code() == nil && !o.SelfDestructed()
}

func (o *ArchiveAccount) SelfDestructed() bool {
	panic("not support SelfDestructed")
}

func (o *ArchiveAccount) GetStorageRoot() common.Hash {
	panic("not support GetStorageRoot")
}

func (o *ArchiveAccount) IsCreated() bool {
	panic("not support IsCreated")
}

func (o *ArchiveAccount) SetCreated(created bool) {
	panic("not support SetCreated")
}
