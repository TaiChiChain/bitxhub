package ledger

/*
#cgo LDFLAGS: /Users/koi/Documents/dev/project/forestore/target/release/libforestore.a  -ldl -lm
#include "/Users/koi/Documents/dev/project/forestore/src/c_ffi/forestore.h"
*/
import "C"
import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"unsafe"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
)

type RustAccount struct {
	stateDBPtr *C.struct_EvmStateDB
	cAddress   C.CAddress
	ethAddress common.Address
	Addr       *types.Address
	created    bool
	changer    *RustStateChanger
	codeCache  *lru.Cache[common.Address, []byte]

	//Logical Account
	logger        logrus.FieldLogger
	originAccount *types.InnerAccount
	dirtyAccount  *types.InnerAccount

	// TODO: use []byte as key
	// The confirmed state of the previous block
	originState map[string][]byte

	// Modified state of previous transactions in the current block
	pendingState map[string][]byte

	// The latest state of the current transaction
	dirtyState map[string][]byte

	originCode []byte
	dirtyCode  []byte

	selfDestructed bool
}

func NewRustAccount(stateDBPtr *C.struct_EvmStateDB, address common.Address, changer *RustStateChanger, codeCache *lru.Cache[common.Address, []byte]) *RustAccount {
	return &RustAccount{
		stateDBPtr: stateDBPtr,
		ethAddress: address,
		Addr:       types.NewAddress(address.Bytes()),
		cAddress:   convertToCAddress(address),
		changer:    changer,
		codeCache:  codeCache,

		logger:         loggers.Logger(loggers.Storage),
		originState:    make(map[string][]byte),
		pendingState:   make(map[string][]byte),
		dirtyState:     make(map[string][]byte),
		selfDestructed: false,
		created:        false,
	}

}

func (o *RustAccount) DeepCopy() IAccount {
	res := &RustAccount{
		logger:         loggers.Logger(loggers.Storage),
		stateDBPtr:     o.stateDBPtr,
		ethAddress:     common.Address(o.Addr.Bytes()),
		Addr:           types.NewAddress(o.Addr.Bytes()),
		cAddress:       convertToCAddress(o.Addr.ETHAddress()),
		created:        o.created,
		codeCache:      o.codeCache,
		changer:        NewChanger(),
		selfDestructed: o.selfDestructed,
	}

	res.originAccount = o.originAccount
	res.dirtyAccount = InnerAccountDeepCopy(o.dirtyAccount)

	res.originState = DeepCopyMapBytes(o.originState)
	res.pendingState = DeepCopyMapBytes(o.pendingState)
	res.dirtyState = DeepCopyMapBytes(o.dirtyState)

	res.originCode = CopyBytes(o.originCode)
	res.dirtyCode = CopyBytes(o.dirtyCode)

	return res
}

func cu256ToUInt256(cu256 C.CU256) *uint256.Int {
	return &uint256.Int{uint64(cu256.low1), uint64(cu256.low2), uint64(cu256.high1), uint64(cu256.high2)}
}

func convertToCU256(u *uint256.Int) C.CU256 {
	return C.CU256{low1: C.uint64_t(u[0]), low2: C.uint64_t(u[1]), high1: C.uint64_t(u[2]), high2: C.uint64_t(u[3])}
}

func convertToCAddress(address common.Address) C.CAddress {
	cAddress := C.CAddress{}
	addrArray := (*[20]C.uint8_t)(unsafe.Pointer(&cAddress.address[0]))
	copy((*[20]uint8)(unsafe.Pointer(addrArray))[:], address[:])
	return cAddress
}

func convertToCH256(hash common.Hash) C.CH256 {
	ch256 := C.CH256{}
	bytesArray := (*[32]C.uint8_t)(unsafe.Pointer(&ch256.bytes[0]))
	copy((*[32]uint8)(unsafe.Pointer(bytesArray))[:], hash[:])
	return ch256
}

func hashKey(key []byte) common.Hash {
	keyHash := sha256.Sum256(key)
	return keyHash
}

func (o *RustAccount) String() string {
	return fmt.Sprintf("{account: %v, code length: %v}", o.ethAddress, len(o.Code()))
}

func (o *RustAccount) initStorageTrie() {
}

func (o *RustAccount) GetAddress() *types.Address {
	return o.Addr
}

// GetState Get state from local cache, if not found, then get it from DB
func (o *RustAccount) GetState(k []byte) (bool, []byte) {

	// GetState From Logical Account
	if value, exist := o.dirtyState[string(k)]; exist {
		o.logger.Debugf("[GetState] get from dirty, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return value != nil, value
	}

	if value, exist := o.pendingState[string(k)]; exist {
		o.logger.Debugf("[GetState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return value != nil, value
	}

	if value, exist := o.originState[string(k)]; exist {
		o.logger.Debugf("[GetState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return value != nil, value
	}

	// GetState From Rust Account
	key := hashKey(k)
	var length C.uintptr_t
	ptr := C.get_any_size_state(o.stateDBPtr, o.cAddress, convertToCH256(key), &length)
	defer C.deallocate_memory(ptr, length)
	valueLen := int(length)
	goSlice := C.GoBytes(unsafe.Pointer(ptr), C.int(length))

	o.originState[string(k)] = goSlice

	return valueLen != 0, goSlice
}

func (o *RustAccount) GetBit256State(key []byte) common.Hash {
	// GetState From Logical Account
	if value, exist := o.dirtyState[string(key)]; exist {
		o.logger.Debugf("[GetState] get from dirty, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return common.BytesToHash(value)
	}

	if value, exist := o.pendingState[string(key)]; exist {
		o.logger.Debugf("[GetState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return common.BytesToHash(value)
	}

	if value, exist := o.originState[string(key)]; exist {
		o.logger.Debugf("[GetState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return common.BytesToHash(value)
	}

	hash_key := hashKey(key)
	cH256 := C.get_state(o.stateDBPtr, o.cAddress, convertToCH256(hash_key))
	goHash := *(*common.Hash)(unsafe.Pointer(&cH256.bytes))

	o.originState[string(key)] = goHash.Bytes()

	return goHash
}

func (o *RustAccount) setBalance(balance *big.Int) {
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.dirtyAccount.Balance = balance
}

func (o *RustAccount) setNonce(nonce uint64) {
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.dirtyAccount.Nonce = nonce
}

func (o *RustAccount) setCodeAndHash(code []byte) {
	ret := crypto.Keccak256Hash(code)
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.dirtyAccount.CodeHash = ret.Bytes()
	o.dirtyCode = code
}

func (o *RustAccount) setState(key []byte, value []byte) {
	o.dirtyState[string(key)] = value
}

func (o *RustAccount) GetCommittedState(k []byte) []byte {
	// GetCommittedState From Logical Account
	if value, exist := o.pendingState[string(k)]; exist {
		if value == nil {
			o.logger.Debugf("[GetCommittedState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
			return (&types.Hash{}).Bytes()
		}
		o.logger.Debugf("[GetCommittedState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return value
	}

	if value, exist := o.originState[string(k)]; exist {
		if value == nil {
			o.logger.Debugf("[GetCommittedState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
			return (&types.Hash{}).Bytes()
		}
		o.logger.Debugf("[GetCommittedState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return value
	}

	key := hashKey(k)
	var length C.uintptr_t
	ptr := C.get_any_size_committed_state(o.stateDBPtr, o.cAddress, convertToCH256(key), &length)
	defer C.deallocate_memory(ptr, length)
	goSlice := C.GoBytes(unsafe.Pointer(ptr), C.int(length))

	o.originState[string(k)] = goSlice

	return goSlice
}

func (o *RustAccount) GetBit256CommittedState(k []byte) common.Hash {
	// GetCommittedState From Logical Account
	if value, exist := o.pendingState[string(k)]; exist {
		if value == nil {
			o.logger.Debugf("[GetCommittedState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
			return common.Hash{}
		}
		o.logger.Debugf("[GetCommittedState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return common.BytesToHash(value)
	}

	if value, exist := o.originState[string(k)]; exist {
		if value == nil {
			o.logger.Debugf("[GetCommittedState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
			return common.Hash{}
		}
		o.logger.Debugf("[GetCommittedState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: value})
		return common.BytesToHash(value)
	}

	key := hashKey(k)
	cH256 := C.get_committed_state(o.stateDBPtr, o.cAddress, convertToCH256(key))
	goHash := *(*common.Hash)(unsafe.Pointer(&cH256.bytes))

	o.originState[string(k)] = goHash.Bytes()

	return goHash
}

// SetState Set account state
func (o *RustAccount) SetState(k []byte, value []byte) {
	_, prev := o.GetState(k)
	o.changer.Append(RustStorageChange{
		account:  o.Addr,
		key:      k,
		prevalue: prev,
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	//o.logger.Debugf("[SetState] addr: %v, key: %v, before state: %v, after state: %v", o.Addr, &bytesLazyLogger{bytes: k}, &bytesLazyLogger{bytes: prev}, &bytesLazyLogger{bytes: value})
	o.setState(k, value)

	// Commit processing
	// key := hashKey(k)
	// valuePtr := (*C.uchar)(unsafe.Pointer(&value[0]))
	// valueLen := C.uintptr_t(len(value))
	// C.set_any_size_state(o.stateDBPtr, o.cAddress, convertToCH256(key), valuePtr, valueLen)
	// runtime.KeepAlive(value)
}

func (o *RustAccount) SetBit256State(k []byte, value common.Hash) {
	_, prev := o.GetState(k)
	o.changer.Append(RustStorageChange{
		account:  o.Addr,
		key:      k,
		prevalue: prev,
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.setState(k, value.Bytes())

	// Commit processing
	// key := hashKey(k)
	// C.set_state(o.stateDBPtr, o.cAddress, convertToCH256(key), convertToCH256(value))
}

// SetCodeAndHash Set the contract code and hash
func (o *RustAccount) SetCodeAndHash(code []byte) {

	ret := crypto.Keccak256Hash(code)
	o.changer.Append(RustCodeChange{
		account:  o.Addr,
		prevcode: o.Code(),
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.logger.Debugf("[SetCodeAndHash] addr: %v, before code hash: %v, after code hash: %v", o.Addr, &bytesLazyLogger{bytes: o.CodeHash()}, &bytesLazyLogger{bytes: ret.Bytes()})
	o.dirtyAccount.CodeHash = ret.Bytes()
	o.dirtyCode = code

	// Commit processing
	// o.codeCache.Add(o.ethAddress, code)
	// codePtr := (*C.uchar)(unsafe.Pointer(&code[0]))
	// codeLen := C.uintptr_t(len(code))
	// C.set_code(o.stateDBPtr, o.cAddress, codePtr, codeLen)
}

// Code return the contract code
func (o *RustAccount) Code() []byte {
	if o.dirtyCode != nil {
		return o.dirtyCode
	}

	if o.originCode != nil {
		return o.originCode
	}

	if bytes.Equal(o.CodeHash(), nil) {
		return nil
	}

	// if v, ok := o.codeCache.Get(o.ethAddress); ok {
	// 	o.originCode = v
	// 	o.dirtyCode = v
	// 	return v
	// }
	var length C.uintptr_t
	ptr := C.get_code(o.stateDBPtr, o.cAddress, &length)
	defer C.deallocate_memory(ptr, length)
	goSlice := C.GoBytes(unsafe.Pointer(ptr), C.int(length))

	o.originCode = goSlice
	o.dirtyCode = goSlice

	return goSlice
}

func (o *RustAccount) CodeHash() []byte {
	if o.dirtyAccount != nil {
		o.logger.Debugf("[CodeHash] get from dirty, addr: %v, code hash: %v", o.Addr, &bytesLazyLogger{bytes: o.dirtyAccount.CodeHash})
		return o.dirtyAccount.CodeHash
	}
	if o.originAccount != nil {
		o.logger.Debugf("[CodeHash] get from origin, addr: %v, code hash: %v", o.Addr, &bytesLazyLogger{bytes: o.originAccount.CodeHash})
		return o.originAccount.CodeHash
	}
	o.logger.Debugf("[CodeHash] not found, addr: %v", o.Addr)

	return nil

	// c_hash := C.get_code_hash(o.stateDBPtr, o.cAddress)
	// goHash := *(*common.Hash)(unsafe.Pointer(&c_hash.bytes))
	// return goHash[:]
}

// SetNonce Set the nonce which indicates the contract number
func (o *RustAccount) SetNonce(nonce uint64) {
	o.changer.Append(RustNonceChange{
		account: o.Addr,
		prev:    o.GetNonce(),
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	if o.originAccount == nil {
		o.logger.Debugf("[SetNonce] addr: %v, before origin nonce: %v, before dirty nonce: %v, after dirty nonce: %v", o.Addr, 0, o.dirtyAccount.Nonce, nonce)
	} else {
		o.logger.Debugf("[SetNonce] addr: %v, before origin nonce: %v, before dirty nonce: %v, after dirty nonce: %v", o.Addr, o.originAccount.Nonce, o.dirtyAccount.Nonce, nonce)
	}
	o.dirtyAccount.Nonce = nonce

	// Commit processing
	// C.set_nonce(o.stateDBPtr, o.cAddress, convertToCU256(uint256.NewInt(nonce)))
}

// GetNonce Get the nonce from user account
func (o *RustAccount) GetNonce() uint64 {
	if o.dirtyAccount != nil {
		o.logger.Debugf("[GetNonce] get from dirty, addr: %v, nonce: %v", o.Addr, o.dirtyAccount.Nonce)
		return o.dirtyAccount.Nonce
	}
	if o.originAccount != nil {
		o.logger.Debugf("[GetNonce] get from origin, addr: %v, nonce: %v", o.Addr, o.originAccount.Nonce)
		return o.originAccount.Nonce
	}

	o.logger.Debugf("[GetNonce] not found, addr: %v, nonce: 0", o.Addr)
	return 0

	// nonce := C.get_nonce(o.stateDBPtr, o.cAddress)
	// return uint64(nonce.low1)
}

// GetBalance Get the balance from the account
func (o *RustAccount) GetBalance() *big.Int {
	if o.dirtyAccount != nil {
		o.logger.Debugf("[GetBalance] get from dirty, addr: %v, balance: %v", o.Addr, o.dirtyAccount.Balance)
		return o.dirtyAccount.Balance
	}
	if o.originAccount != nil {
		o.logger.Debugf("[GetBalance] get from origin, addr: %v, balance: %v", o.Addr, o.originAccount.Balance)
		return o.originAccount.Balance
	}

	o.logger.Debugf("[GetBalance] not found, addr: %v, balance: 0", o.Addr)
	return new(big.Int).SetInt64(0)

	// balance := C.get_balance(o.stateDBPtr, o.cAddress)
	// return cu256ToUInt256(balance).ToBig()
}

// SetBalance Set the balance to the account
func (o *RustAccount) SetBalance(balance *big.Int) {
	o.changer.Append(RustBalanceChange{
		account: o.Addr,
		prev:    new(big.Int).Set(o.GetBalance()),
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	if o.originAccount == nil {
		o.logger.Debugf("[SetBalance] addr: %v, before origin balance: %v, before dirty balance: %v, after dirty balance: %v", o.Addr, 0, o.dirtyAccount.Balance, balance)
	} else {
		o.logger.Debugf("[SetBalance] addr: %v, before origin balance: %v, before dirty balance: %v, after dirty balance: %v", o.Addr, o.originAccount.Balance, o.dirtyAccount.Balance, balance)
	}
	o.dirtyAccount.Balance = balance

	// Commit processing
	// u, _ := uint256.FromBig(balance)
	// C.set_balance(o.stateDBPtr, o.cAddress, convertToCU256(u))

}

func (o *RustAccount) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	o.logger.Debugf("[SubBalance] addr: %v, sub amount: %v", o.Addr, amount)
	o.SetBalance(new(big.Int).Sub(o.GetBalance(), amount))
}

func (o *RustAccount) AddBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	o.logger.Debugf("[AddBalance] addr: %v, add amount: %v", o.Addr, amount)
	o.SetBalance(new(big.Int).Add(o.GetBalance(), amount))

}

// Finalise moves all dirty states into the pending states.
// Return all dirty state keys
func (o *RustAccount) Finalise() [][]byte {
	//keys2Preload := make([][]byte, 0, len(o.dirtyState))
	for key, value := range o.dirtyState {
		o.pendingState[key] = value

		// collect all storage key of the account
		//keys2Preload = append(keys2Preload, utils.CompositeStorageKey(o.Addr, []byte(key)))
	}
	o.dirtyState = make(map[string][]byte)
	return nil
}

func (o *RustAccount) SetSelfDestructed(selfDestructed bool) {
	o.selfDestructed = selfDestructed
	// Commit processing
	// C.self_destruct(o.stateDBPtr, o.cAddress)
}

func (o *RustAccount) IsEmpty() bool {
	return o.GetBalance().Sign() == 0 && o.GetNonce() == 0 && o.Code() == nil && !o.SelfDestructed()
}

func (o *RustAccount) SelfDestructed() bool {
	return o.selfDestructed
}

func (o *RustAccount) GetStorageRoot() common.Hash {
	panic("not support")
}

func (o *RustAccount) IsCreated() bool {
	return o.created
}

func (o *RustAccount) SetCreated(created bool) {
	o.created = created
}

func (o *RustAccount) getStateJournal() map[string][]byte {
	prevStates := make(map[string][]byte)

	for key, value := range o.pendingState {
		origVal := o.originState[key]
		if !bytes.Equal(origVal, value) {
			prevStates[key] = origVal
		}
	}
	return prevStates
}
