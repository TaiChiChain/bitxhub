package ledger

/*
#cgo LDFLAGS: "/Users/koi/Documents/dev/project/forestore/target/release/libforestore.a"  -ldl -lm
#include "/Users/koi/Documents/dev/project/forestore/src/c_ffi/forestore.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"path"
	"runtime"
	"sort"
	"unsafe"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/blockstm"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
	etherTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
)

type RustStateLedger struct {
	logger logrus.FieldLogger

	stateDBPtr     *C.struct_EvmStateDB
	stateDBViewPtr *C.struct_EvmStateDB

	Accounts    map[string]IAccount
	repo        *repo.Repo
	blockHeight uint64
	thash       *types.Hash
	txIndex     int

	validRevisions []revision
	nextRevisionId int
	changer        *RustStateChanger

	AccessList *AccessList
	preimages  map[types.Hash][]byte
	Refund     uint64
	Logs       *evmLogs

	transientStorage transientStorage

	codeCache *lru.Cache[ethcommon.Address, []byte]

	// Block-stm related fields
	mvHashmap    *blockstm.MVHashMap
	incarnation  int
	readMap      map[blockstm.Key]blockstm.ReadDescriptor
	writeMap     map[blockstm.Key]blockstm.WriteDescriptor
	revertedKeys map[blockstm.Key]struct{}
	dep          int
} /**/

func NewRustStateLedger(rep *repo.Repo, height uint64) *RustStateLedger {
	dir := repo.GetStoragePath(rep.RepoRoot, storagemgr.Rust_Ledger)
	logDir := path.Join(dir, "logs")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		panic(err)
	}
	C.set_up_default_logger(C.CString(logDir))

	version := height
	initialVersion := rep.GenesisConfig.EpochInfo.StartBlock
	if initialVersion != 0 && initialVersion != 1 {
		panic("Genesis start block num should be 0 or 1")
	}
	C.rollback_evm_state_db(C.CString(dir), C.uint64_t(version))
	cOpts := C.CEvmStateDBOptions{genesis_version: C.uint64_t(initialVersion), snapshot_rewrite_interval: C.uint64_t(rep.Config.Forestore.SnapshotRewriteInterval)}
	stateDbPtr := C.new_evm_state_db(C.CString(dir), cOpts)
	stateDBViewPtr := C.evm_state_db_view(stateDbPtr)
	codeCache, _ := lru.New[ethcommon.Address, []byte](1000)
	return &RustStateLedger{
		logger: loggers.Logger(loggers.Storage),

		repo:           rep,
		stateDBPtr:     stateDbPtr,
		stateDBViewPtr: stateDBViewPtr,
		Accounts:       make(map[string]IAccount),
		Logs:           newEvmLogs(),

		preimages:        make(map[types.Hash][]byte),
		transientStorage: newTransientStorage(),
		changer:          NewChanger(),

		AccessList: NewAccessList(),
		codeCache:  codeCache,

		revertedKeys: make(map[blockstm.Key]struct{}),
	}
}

func (r *RustStateLedger) Copy() StateLedger {
	codeCache, _ := lru.New[ethcommon.Address, []byte](1000)
	lg := &RustStateLedger{
		repo:           r.repo,
		stateDBPtr:     r.stateDBPtr,
		stateDBViewPtr: r.stateDBViewPtr,
		blockHeight:    r.blockHeight,
		Accounts:       make(map[string]IAccount),
		Logs:           newEvmLogs(),

		preimages:        make(map[types.Hash][]byte),
		transientStorage: newTransientStorage(),
		changer:          NewChanger(),

		AccessList:   NewAccessList(),
		codeCache:    codeCache,
		revertedKeys: make(map[blockstm.Key]struct{}),
	}

	return lg
}

func (l *RustStateLedger) SetFromOrigin(origin StateLedger) {
	rustOrigin := origin.(*RustStateLedger)
	l.repo = rustOrigin.repo
	l.stateDBPtr = rustOrigin.stateDBPtr
	l.stateDBViewPtr = rustOrigin.stateDBViewPtr

	l.Logs = rustOrigin.Logs
	l.blockHeight = rustOrigin.blockHeight
	l.codeCache = rustOrigin.codeCache
}

func (l *RustStateLedger) Reset() {
	l.preimages = make(map[types.Hash][]byte)
	l.changer = NewChanger()
	l.AccessList = NewAccessList()
	l.Logs = newEvmLogs()
	l.revertedKeys = make(map[blockstm.Key]struct{})
	l.thash = nil
	l.txIndex = 0
	l.validRevisions = make([]revision, 0)
	l.nextRevisionId = 0
	l.Refund = 0
	l.transientStorage = newTransientStorage()
	l.mvHashmap = nil
	l.dep = 0
	l.incarnation = 0
	l.readMap = nil
	l.writeMap = nil
	codeCache, _ := lru.New[ethcommon.Address, []byte](1000)
	l.codeCache = codeCache
}

func (r *RustStateLedger) GetOrCreateAccount(address *types.Address) IAccount {
	account := r.GetAccount(address)
	if account != nil {
		return account
	} else {
		rustAccount := NewRustAccount(r.stateDBPtr, address.ETHAddress(), r.changer, r.codeCache)
		rustAccount.SetCreated(true)
		r.changer.Append(RustCreateObjectChange{account: address})

		r.Accounts[address.String()] = rustAccount
		mvwrite(r, blockstm.NewAddressKey(address))
		return rustAccount
	}

}

func (r *RustStateLedger) GetAccount(address *types.Address) IAccount {
	return mvread(r, blockstm.NewAddressKey(address), nil, func(r *RustStateLedger) IAccount {
		addr := address.String()

		value, ok := r.Accounts[addr]
		if ok {
			return value
		}
		cAddress := convertToCAddress(address.ETHAddress())
		exist := bool(C.exist(r.stateDBPtr, cAddress))
		if exist {
			account := NewRustAccount(r.stateDBPtr, address.ETHAddress(), r.changer, r.codeCache)
			account.CodeHash()
			c_hash := C.get_code_hash(r.stateDBPtr, cAddress)
			goHash := *(*ethcommon.Hash)(unsafe.Pointer(&c_hash.bytes))
			account.originAccount = &types.InnerAccount{
				Nonce:   uint64(C.get_nonce(r.stateDBPtr, cAddress).low1),
				Balance: cu256ToUInt256(C.get_balance(r.stateDBPtr, cAddress)).ToBig(),
			}
			isHashNotZero := goHash != ethcommon.Hash{}
			if isHashNotZero {
				account.originAccount.CodeHash = goHash[:]
			}
			if !bytes.Equal(account.originAccount.CodeHash, nil) {
				var length C.uintptr_t
				ptr := C.get_code(r.stateDBPtr, cAddress, &length)
				defer C.deallocate_memory(ptr, length)
				goSlice := C.GoBytes(unsafe.Pointer(ptr), C.int(length))

				account.originCode = goSlice
				account.dirtyCode = goSlice
			}

			r.Accounts[addr] = account
			return account
		} else {
			return nil
		}
	})
}

func (r *RustStateLedger) GetBalance(address *types.Address) *big.Int {
	return mvread(r, blockstm.NewSubpathKey(address, BalancePath), ethcommon.Big0, func(r *RustStateLedger) *big.Int {
		account := r.GetAccount(address)
		if account != nil {
			return account.GetBalance()
		}

		return ethcommon.Big0
	})
}

func (r *RustStateLedger) SetBalance(address *types.Address, balance *big.Int) {
	account := r.GetOrCreateAccount(address)
	if r.mvHashmap != nil {
		// ensure a read balance operation is recorded in mvHashmap
		r.GetBalance(address)
	}

	account = r.mvRecordWritten(account)
	account.SetBalance(balance)
	mvwrite(r, blockstm.NewSubpathKey(address, BalancePath))
}

func (r *RustStateLedger) SubBalance(address *types.Address, balance *big.Int) {
	account := r.GetOrCreateAccount(address)
	if r.mvHashmap != nil {
		// ensure a read balance operation is recorded in mvHashmap
		r.GetBalance(address)
	}
	if account != nil {
		account = r.mvRecordWritten(account)
		account.SubBalance(balance)
		mvwrite(r, blockstm.NewSubpathKey(address, BalancePath))
	}
}

func (r *RustStateLedger) AddBalance(address *types.Address, balance *big.Int) {
	account := r.GetOrCreateAccount(address)
	if r.mvHashmap != nil {
		// ensure a read balance operation is recorded in mvHashmap
		r.GetBalance(address)
	}
	account = r.mvRecordWritten(account)
	account.AddBalance(balance)
	mvwrite(r, blockstm.NewSubpathKey(address, BalancePath))
}

func (r *RustStateLedger) GetState(address *types.Address, bytes []byte) (bool, []byte) {
	return mvread2(r, blockstm.NewStateKey(address, *types.NewHash(bytes)), true, (&types.Hash{}).Bytes(), func(r *RustStateLedger) (bool, []byte) {
		account := r.GetAccount(address)
		if account != nil {
			return account.GetState(bytes)
		}
		return false, []byte{}
	})
}

func (r *RustStateLedger) GetBit256State(address *types.Address, bytes []byte) (bool, []byte) {
	return mvread2(r, blockstm.NewStateKey(address, *types.NewHash(bytes)), true, (&types.Hash{}).Bytes(), func(r *RustStateLedger) (bool, []byte) {
		account := r.GetAccount(address)
		if account != nil {
			return account.(*RustAccount).GetBit256State(bytes)
		}
		return false, []byte{}
	})
}

func (r *RustStateLedger) SetState(address *types.Address, key []byte, value []byte) {
	account := r.GetOrCreateAccount(address)

	account = r.mvRecordWritten(account)
	account.SetState(key, value)
	mvwrite(r, blockstm.NewStateKey(address, *types.NewHash(key)))
}

func (r *RustStateLedger) SetBit256State(address *types.Address, key []byte, value ethcommon.Hash) {
	account := r.GetOrCreateAccount(address)

	account = r.mvRecordWritten(account)
	account.(*RustAccount).SetBit256State(key, value)
	mvwrite(r, blockstm.NewStateKey(address, *types.NewHash(key)))
}

func (r *RustStateLedger) SetCode(address *types.Address, bytes []byte) {
	account := r.GetOrCreateAccount(address)

	account = r.mvRecordWritten(account)
	account.SetCodeAndHash(bytes)
	mvwrite(r, blockstm.NewSubpathKey(address, CodePath))
}

func (r *RustStateLedger) GetCode(address *types.Address) []byte {
	return mvread(r, blockstm.NewSubpathKey(address, CodePath), nil, func(r *RustStateLedger) []byte {
		account := r.GetAccount(address)
		if account != nil {
			return account.Code()
		}
		return nil
	})
}

func (r *RustStateLedger) SetNonce(address *types.Address, u uint64) {
	account := r.GetOrCreateAccount(address)

	account = r.mvRecordWritten(account)
	account.SetNonce(u)
	mvwrite(r, blockstm.NewSubpathKey(address, NoncePath))
}

func (r *RustStateLedger) GetNonce(address *types.Address) uint64 {
	return mvread(r, blockstm.NewSubpathKey(address, NoncePath), 0, func(l *RustStateLedger) uint64 {
		account := l.GetAccount(address)
		if account != nil {
			return account.GetNonce()
		}
		return 0
	})
}

func (r *RustStateLedger) GetCodeHash(address *types.Address) *types.Hash {
	return mvread(r, blockstm.NewSubpathKey(address, CodePath), &types.Hash{}, func(r *RustStateLedger) *types.Hash {
		account := r.GetAccount(address)
		if account != nil && !account.IsEmpty() {
			return types.NewHash(account.CodeHash())
		}
		return &types.Hash{}

	})
}

func (r *RustStateLedger) GetCodeSize(address *types.Address) int {
	return mvread(r, blockstm.NewSubpathKey(address, CodePath), 0, func(l *RustStateLedger) int {
		account := l.GetAccount(address)
		if account != nil && !account.IsEmpty() {
			if code := account.Code(); code != nil {
				return len(code)
			}
		}
		return 0
	})
}

func (r *RustStateLedger) AddRefund(gas uint64) {
	r.changer.Append(RustRefundChange{prev: r.Refund})
	r.Refund += gas

	// Commit processing
	//C.add_refund(r.stateDBPtr, C.uint64_t(gas))
}

func (r *RustStateLedger) SubRefund(gas uint64) {
	r.changer.Append(RustRefundChange{prev: r.Refund})
	if gas > r.Refund {
		panic(fmt.Sprintf("Refund counter below zero (gas: %d > refund: %d)", gas, r.Refund))
	}
	r.Refund -= gas
	// Commit processing
	//C.sub_refund(r.stateDBPtr, C.uint64_t(gas))
}

func (r *RustStateLedger) GetRefund() uint64 {
	return r.Refund
}

func (r *RustStateLedger) GetCommittedState(address *types.Address, bytes []byte) []byte {
	return mvread(r, blockstm.NewStateKey(address, *types.NewHash(bytes)), nil, func(r *RustStateLedger) []byte {
		account := r.GetAccount(address)
		if account != nil {
			return account.GetCommittedState(bytes)
		}
		return (&types.Hash{}).Bytes()
	})

}

func (r *RustStateLedger) GetBit256CommittedState(address *types.Address, bytes []byte) ethcommon.Hash {
	return mvread(r, blockstm.NewStateKey(address, *types.NewHash(bytes)), ethcommon.Hash{}, func(r *RustStateLedger) ethcommon.Hash {
		account := r.GetAccount(address)
		if account != nil {
			return account.(*RustAccount).GetBit256CommittedState(bytes)
		}
		return ethcommon.Hash{}
	})
}

func (r *RustStateLedger) Commit() (*types.Hash, error) {
	accounts := r.collectDirtyData()
	for _, acc := range accounts {
		account := acc.(*RustAccount)

		// update balance
		u, _ := uint256.FromBig(account.dirtyAccount.Balance)
		C.set_balance(account.stateDBPtr, account.cAddress, convertToCU256(u))

		// update nonce
		C.set_nonce(account.stateDBPtr, account.cAddress, convertToCU256(uint256.NewInt(account.dirtyAccount.Nonce)))

		// update state
		for k, value := range account.pendingState {
			key := hashKey([]byte(k))
			valuePtr := (*C.uchar)(unsafe.Pointer(&value[0]))
			valueLen := C.uintptr_t(len(value))
			C.set_any_size_state(account.stateDBPtr, account.cAddress, convertToCH256(key), valuePtr, valueLen)
			runtime.KeepAlive(value)
		}

		// update code
		if !bytes.Equal(account.originCode, account.dirtyCode) && account.dirtyCode != nil {
			code := account.dirtyCode
			//account.codeCache.Add(account.ethAddress, code)
			codePtr := (*C.uchar)(unsafe.Pointer(&code[0]))
			codeLen := C.uintptr_t(len(code))
			C.set_code(account.stateDBPtr, account.cAddress, codePtr, codeLen)
		}
		if account.selfDestructed {
			C.self_destruct(account.stateDBPtr, account.cAddress)
		}

	}
	committedRes := C.commit(r.stateDBPtr)
	rootHash := *(*[32]byte)(unsafe.Pointer(&committedRes.root_hash.bytes))
	committed_height := committedRes.committed_version
	if r.blockHeight != uint64(committed_height) {
		panic("committed height not match")
	}
	r.Accounts = make(map[string]IAccount)
	return types.NewHash(rootHash[:]), nil
}

func (r *RustStateLedger) SelfDestruct(address *types.Address) bool {
	account := r.GetAccount(address)
	if account == nil {
		return false
	}
	account.SetSelfDestructed(true)
	account.SetBalance(new(big.Int))

	mvwrite(r, blockstm.NewSubpathKey(address, SuicidePath))
	mvwrite(r, blockstm.NewSubpathKey(address, BalancePath))
	return true
}

func (r *RustStateLedger) HasSelfDestructed(address *types.Address) bool {

	return mvread(r, blockstm.NewSubpathKey(address, SuicidePath), false, func(l *RustStateLedger) bool {
		account := r.GetAccount(address)
		if account != nil {
			return account.SelfDestructed()
		}
		return false
	})
}

func (r *RustStateLedger) Selfdestruct6780(address *types.Address) {
	account := r.GetAccount(address)
	if account == nil {
		return
	}

	if account.IsCreated() {
		r.SelfDestruct(address)
	}
}

func (r *RustStateLedger) Exist(address *types.Address) bool {
	exist := r.GetAccount(address) != nil
	return exist
}

func (r *RustStateLedger) Empty(address *types.Address) bool {
	account := r.GetAccount(address)
	empty := account == nil || account.IsEmpty()
	return empty
}

func (r *RustStateLedger) AddressInAccessList(addr types.Address) bool {
	return r.AccessList.ContainsAddress(addr)
}

func (r *RustStateLedger) SlotInAccessList(addr types.Address, slot types.Hash) (bool, bool) {
	return r.AccessList.Contains(addr, slot)
}

func (r *RustStateLedger) AddAddressToAccessList(addr types.Address) {
	if r.AccessList.AddAddress(addr) {
		r.changer.Append(RustAccessListAddAccountChange{address: &addr})
	}
}

func (r *RustStateLedger) AddSlotToAccessList(addr types.Address, slot types.Hash) {
	addrMod, slotMod := r.AccessList.AddSlot(addr, slot)
	if addrMod {
		r.changer.Append(RustAccessListAddAccountChange{address: &addr})
	}
	if slotMod {
		r.changer.Append(RustAccessListAddSlotChange{
			address: &addr,
			slot:    &slot,
		})
	}
}

func (r *RustStateLedger) Prepare(rules params.Rules, sender, coinbase ethcommon.Address, dst *ethcommon.Address, precompiles []ethcommon.Address, list etherTypes.AccessList) {
	r.AccessList = NewAccessList()
	if rules.IsBerlin {
		// Clear out any leftover from previous executions
		al := NewAccessList()
		r.AccessList = al

		al.AddAddress(*types.NewAddress(sender.Bytes()))
		if dst != nil {
			al.AddAddress(*types.NewAddress(dst.Bytes()))
			// If it's a create-tx, the destination will be added inside evm.create
		}
		for _, addr := range precompiles {
			al.AddAddress(*types.NewAddress(addr.Bytes()))
		}
		for _, el := range list {
			al.AddAddress(*types.NewAddress(el.Address.Bytes()))
			for _, key := range el.StorageKeys {
				al.AddSlot(*types.NewAddress(el.Address.Bytes()), *types.NewHash(key.Bytes()))
			}
		}
		// if rules.IsShanghai { // EIP-3651: warm coinbase
		// 	al.AddAddress(coinbase)
		// }
	}
	// Reset transient storage at the beginning of transaction execution
	r.transientStorage = newTransientStorage()
}

func (r *RustStateLedger) Preimages() map[types.Hash][]byte {
	return r.preimages
}

func (r *RustStateLedger) AddPreimage(address types.Hash, preimage []byte) {

	if _, ok := r.preimages[address]; !ok {
		r.changer.Append(RustAddPreimageChange{hash: address})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		r.preimages[address] = pi
	}

	// Commit processing
	// preimagePtr := (*C.uchar)(unsafe.Pointer(&preimage[0]))
	// preimageLen := C.uintptr_t(len(preimage))
	// C.add_preimage(r.stateDBPtr, convertToCH256(address.ETHHash()), preimagePtr, preimageLen)
}

func (r *RustStateLedger) SetTxContext(thash *types.Hash, ti int) {
	r.thash = thash
	r.txIndex = ti
}

func (r *RustStateLedger) Clear() {
	r.Accounts = make(map[string]IAccount)
}

func (r *RustStateLedger) RevertToSnapshot(snapshot int) {
	idx := sort.Search(len(r.validRevisions), func(i int) bool {
		return r.validRevisions[i].id >= snapshot
	})
	if idx == len(r.validRevisions) || r.validRevisions[idx].id != snapshot {
		panic(fmt.Errorf("revision id %v cannod be reverted", snapshot))
	}
	snap := r.validRevisions[idx].changerIndex

	r.changer.Revert(r, snap)
	r.validRevisions = r.validRevisions[:idx]
	C.revert_to_snapshot(r.stateDBPtr, C.uintptr_t(snapshot))
}

func (r *RustStateLedger) Snapshot() int {
	id := r.nextRevisionId
	r.nextRevisionId++
	r.validRevisions = append(r.validRevisions, revision{id: id, changerIndex: r.changer.length()})
	C.snapshot(r.stateDBPtr)
	return id
}

func (r *RustStateLedger) setTransientState(addr types.Address, key, value ethcommon.Hash) {
	C.set_transient_state(r.stateDBPtr, convertToCAddress(addr.ETHAddress()), convertToCH256(key), convertToCH256(value))
}

func (r *RustStateLedger) AddLog(log *types.EvmLog) {
	if log.TransactionHash == nil {
		log.TransactionHash = r.thash
	}

	log.TransactionIndex = uint64(r.txIndex)
	r.changer.Append(RustAddLogChange{txHash: log.TransactionHash})
	log.LogIndex = uint64(r.Logs.LogSize)
	if _, ok := r.Logs.Logs[*log.TransactionHash]; !ok {
		r.Logs.Logs[*log.TransactionHash] = make([]*types.EvmLog, 0)
	}

	r.Logs.Logs[*log.TransactionHash] = append(r.Logs.Logs[*log.TransactionHash], log)
	r.Logs.LogSize++
}

func (r *RustStateLedger) GetLogs(txHash types.Hash, height uint64) []*types.EvmLog {
	logs := r.Logs.Logs[txHash]
	for _, l := range logs {
		l.BlockNumber = height
	}
	return logs
}

func (r *RustStateLedger) RollbackState(height uint64, lastStateRoot *types.Hash) error {
	// TODO implement me
	//panic("implement me")
	return nil
}

func (r *RustStateLedger) PrepareBlock(lastStateRoot *types.Hash, currentExecutingHeight uint64) {
	r.Logs = newEvmLogs()
	// TODO lastStateRoot relate logic
	r.blockHeight = currentExecutingHeight
}

func (r *RustStateLedger) PrepareTranct() {
	C.prepare_transact(r.stateDBPtr)
}

func (r *RustStateLedger) FinalizeTransact() {
	C.finalize_transact(r.stateDBPtr)
}

func (r *RustStateLedger) RollbackTransact() {
	C.rollback_transact(r.stateDBPtr)
}

func (r *RustStateLedger) ClearChangerAndRefund() {
	r.Refund = 0
	r.changer.Reset()
	r.validRevisions = r.validRevisions[:0]
	r.nextRevisionId = 0
}

func (r *RustStateLedger) Close() {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) Finalise() {
	for _, account := range r.Accounts {
		account.Finalise()

		account.SetCreated(false)
	}

	r.ClearChangerAndRefund()
}

func (r *RustStateLedger) Version() uint64 {
	return r.blockHeight
}

func (r *RustStateLedger) NewView(blockHeader *types.BlockHeader, enableSnapshot bool) (StateLedger, error) {
	return &RustStateLedger{
		repo:           r.repo,
		stateDBPtr:     r.stateDBViewPtr,
		stateDBViewPtr: r.stateDBViewPtr,
		Accounts:       make(map[string]IAccount),
		changer:        NewChanger(),
		AccessList:     NewAccessList(),
		Logs:           newEvmLogs(),
		blockHeight:    blockHeader.Number,
	}, nil
}

func (r *RustStateLedger) IterateTrie(snapshotMeta *SnapshotMeta, kv kv.Storage, errC chan error) {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) GetTrieSnapshotMeta() (*SnapshotMeta, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) VerifyTrie(blockHeader *types.BlockHeader) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) Prove(rootHash ethcommon.Hash, key []byte) (*jmt.ProofResult, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) GenerateSnapshot(blockHeader *types.BlockHeader, errC chan error) {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) GetHistoryRange() (uint64, uint64) {
	//TODO implement me
	panic("implement me")
}

func (r *RustStateLedger) CurrentBlockHeight() uint64 {
	return r.blockHeight
}

func (r *RustStateLedger) GetStateDelta(blockNumber uint64) *types.StateDelta {
	//TODO implement me
	panic("implement me")
}

var _ StateLedger = (*RustStateLedger)(nil)

func (r *RustStateLedger) setAccount(account IAccount) {
	r.Accounts[account.GetAddress().String()] = account
	//r.logger.Debugf("[Revert setAccount] addr: %v, account: %v", account.GetAddress(), account)
}

func (r *RustStateLedger) SetMVHashmap(mvhm *blockstm.MVHashMap) {
	r.mvHashmap = mvhm
	r.dep = -1
}

func (r *RustStateLedger) GetMVHashmap() *blockstm.MVHashMap {
	return r.mvHashmap
}

func (r *RustStateLedger) MVWriteList() []blockstm.WriteDescriptor {
	writes := make([]blockstm.WriteDescriptor, 0, len(r.writeMap))

	for _, v := range r.writeMap {
		if _, ok := r.revertedKeys[v.Path]; !ok {
			writes = append(writes, v)
		}
	}

	return writes
}

func (r *RustStateLedger) MVFullWriteList() []blockstm.WriteDescriptor {
	writes := make([]blockstm.WriteDescriptor, 0, len(r.writeMap))

	for _, v := range r.writeMap {
		writes = append(writes, v)
	}

	return writes
}

func (r *RustStateLedger) MVReadMap() map[blockstm.Key]blockstm.ReadDescriptor {
	return r.readMap
}

func (r *RustStateLedger) MVReadList() []blockstm.ReadDescriptor {
	reads := make([]blockstm.ReadDescriptor, 0, len(r.readMap))

	for _, v := range r.MVReadMap() {
		reads = append(reads, v)
	}

	return reads
}

func (r *RustStateLedger) ensureReadMap() {
	if r.readMap == nil {
		r.readMap = make(map[blockstm.Key]blockstm.ReadDescriptor)
	}
}

func (r *RustStateLedger) ensureWriteMap() {
	if r.writeMap == nil {
		r.writeMap = make(map[blockstm.Key]blockstm.WriteDescriptor)
	}
}

func (r *RustStateLedger) HadInvalidRead() bool {
	return r.dep >= 0
}

func (r *RustStateLedger) DepTxIndex() int {
	return r.dep
}

func (r *RustStateLedger) SetIncarnation(inc int) {
	r.incarnation = inc
}

func mvread[T any](r *RustStateLedger, k blockstm.Key, defaultV T, readStorage func(r *RustStateLedger) T) (v T) {
	if r.mvHashmap == nil {
		return readStorage(r)
	}

	r.ensureReadMap()

	if r.writeMap != nil {
		if _, ok := r.writeMap[k]; ok {
			return readStorage(r)
		}
	}

	if !k.IsAddress() {
		// If we are reading subpath from a deleted account, return default value instead of reading from MVHashmap
		addr := k.GetAddress()
		if r.GetAccount(addr) == nil {
			return defaultV
		}
	}

	res := r.mvHashmap.Read(k, r.txIndex)

	var rd blockstm.ReadDescriptor

	rd.V = blockstm.Version{
		TxnIndex:    res.DepIdx(),
		Incarnation: res.Incarnation(),
	}

	rd.Path = k

	switch res.Status() {
	case blockstm.MVReadResultDone:
		{
			v = readStorage(res.Value().(*RustStateLedger))
			rd.Kind = blockstm.ReadKindMap
		}
	case blockstm.MVReadResultDependency:
		{
			r.dep = res.DepIdx()

			panic("Found dependency")
		}
	case blockstm.MVReadResultNone:
		{
			v = readStorage(r)
			rd.Kind = blockstm.ReadKindStorage
		}
	default:
		return defaultV
	}

	// TODO: I assume we don't want to overwrite an existing read because this could - for example - change a storage
	//  read to map if the same value is read multiple times.
	if _, ok := r.readMap[k]; !ok {
		r.readMap[k] = rd
	}

	return
}

func mvread2[T any, V any](r *RustStateLedger, k blockstm.Key, defaultV T, defaultV2 V, readStorage2 func(r *RustStateLedger) (T, V)) (v T, v2 V) {
	if r.mvHashmap == nil {
		return readStorage2(r)
	}

	r.ensureReadMap()

	if r.writeMap != nil {
		if _, ok := r.writeMap[k]; ok {
			return readStorage2(r)
		}
	}

	if !k.IsAddress() {
		// If we are reading subpath from a deleted account, return default value instead of reading from MVHashmap
		addr := k.GetAddress()
		if r.GetAccount(addr) == nil {
			return defaultV, defaultV2
		}
	}

	res := r.mvHashmap.Read(k, r.txIndex)

	var rd blockstm.ReadDescriptor

	rd.V = blockstm.Version{
		TxnIndex:    res.DepIdx(),
		Incarnation: res.Incarnation(),
	}

	rd.Path = k

	switch res.Status() {
	case blockstm.MVReadResultDone:
		{
			v, v2 = readStorage2(res.Value().(*RustStateLedger))
			rd.Kind = blockstm.ReadKindMap
		}
	case blockstm.MVReadResultDependency:
		{
			r.dep = res.DepIdx()

			panic("Found dependency")
		}
	case blockstm.MVReadResultNone:
		{
			v, v2 = readStorage2(r)
			rd.Kind = blockstm.ReadKindStorage
		}
	default:
		return defaultV, defaultV2
	}

	// TODO: I assume we don't want to overwrite an existing read because this could - for example - change a storage
	//  read to map if the same value is read multiple times.
	if _, ok := r.readMap[k]; !ok {
		r.readMap[k] = rd
	}

	return
}

func mvwrite(r *RustStateLedger, k blockstm.Key) {
	if r.mvHashmap != nil {
		r.ensureWriteMap()
		r.writeMap[k] = blockstm.WriteDescriptor{
			Path: k,
			V:    r.blockStmVersion(),
			Val:  r,
		}
	}
}

func RustRevertWrite(r *RustStateLedger, k blockstm.Key) {
	r.revertedKeys[k] = struct{}{}
}

func mvwritten(r *RustStateLedger, k blockstm.Key) bool {
	if r.mvHashmap == nil || r.writeMap == nil {
		return false
	}

	_, ok := r.writeMap[k]

	return ok
}

func (r *RustStateLedger) FlushMVWriteSet() {
	if r.mvHashmap != nil && r.writeMap != nil {
		r.mvHashmap.FlushMVWriteSet(r.MVFullWriteList())
	}
}

func (r *RustStateLedger) ApplyMVWriteSet(writes []blockstm.WriteDescriptor) {
	for i := range writes {
		path := writes[i].Path
		sr := writes[i].Val.(*RustStateLedger)

		if path.IsState() {
			addr := path.GetAddress()
			stateKey := path.GetStateKey()
			_, state := sr.GetState(addr, stateKey.Bytes())
			r.SetState(addr, stateKey.Bytes(), state)
		} else if path.IsAddress() {
			continue
		} else {
			addr := path.GetAddress()

			switch path.GetSubpath() {
			case BalancePath:
				r.SetBalance(addr, sr.GetBalance(addr))
			case NoncePath:
				r.SetNonce(addr, sr.GetNonce(addr))
			case CodePath:
				r.SetCode(addr, sr.GetCode(addr))
			case SuicidePath:
				stateObject := sr.GetAccount(addr)
				if stateObject != nil && stateObject.SelfDestructed() {
					r.SelfDestruct(addr)
				}
			default:
				panic(fmt.Errorf("unknown key type: %d", path.GetSubpath()))
			}
		}
	}
}

func (r *RustStateLedger) blockStmVersion() blockstm.Version {
	return blockstm.Version{
		TxnIndex:    r.txIndex,
		Incarnation: r.incarnation,
	}
}

func (r *RustStateLedger) mvRecordWritten(object IAccount) IAccount {
	if r.mvHashmap == nil {
		return object
	}

	addrKey := blockstm.NewAddressKey(object.GetAddress())

	if mvwritten(r, addrKey) {
		return object
	}

	// todo
	// check
	r.Accounts[object.GetAddress().String()] = object.(*RustAccount).DeepCopy()
	mvwrite(r, addrKey)

	return r.Accounts[object.GetAddress().String()]
}

func (r *RustStateLedger) collectDirtyData() map[ethcommon.Address]IAccount {
	dirtyAccounts := make(map[ethcommon.Address]IAccount)
	for _, acc := range r.Accounts {
		account := acc.(*RustAccount)

		prevStates := account.getStateJournal()
		if account.originAccount.InnerAccountChanged(account.dirtyAccount) || len(prevStates) != 0 || account.selfDestructed {
			dirtyAccounts[account.ethAddress] = account
		}

		if account.originCode == nil && account.originAccount != nil && account.originAccount.CodeHash != nil {
			if v, ok := account.codeCache.Get(account.ethAddress); ok {
				account.originCode = v
			} else {
				var length C.uintptr_t
				ptr := C.get_code(account.stateDBPtr, account.cAddress, &length)
				defer C.deallocate_memory(ptr, length)
				goSlice := C.GoBytes(unsafe.Pointer(ptr), C.int(length))
				account.originCode = goSlice
			}
		}
	}
	r.Clear()
	return dirtyAccounts
}
