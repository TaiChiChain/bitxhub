package ledger

/*
#cgo LDFLAGS: /home/bcds/krc/forestore/crates/history_store/target/release/libhistory_store.a -lm -lstdc++
#include "/home/bcds/krc/forestore/crates/history_store/src/c_ffi/history.h"
#include <stdlib.h>
*/
import "C"
import (
	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"math/big"
	"os"
	"path"
	"unsafe"
)

// #cgo LDFLAGS: /Users/zhangqirui/workplace/forestore/crates/history_store/target/release/libhistory_store.a -ldl -lm -lc++ -lc++abi
// #include "/Users/zhangqirui/workplace/forestore/crates/history_store/src/c_ffi/history.h"
// #include <stdlib.h>

type ArchiveStateLedger struct {
	repo   *repo.Repo
	logger logrus.FieldLogger

	archiveStateDB *C.EvmArchiveStateLedger
	blockHeight    uint64
	rootHash       common.Hash
}

func ConvertToCAccountKey(address *types.Address) C.CAccountKey {
	accountKey := utils.CompositeAccountKey(address)
	cAccountKey := C.CAccountKey{}
	cArray := (*[40]C.uint8_t)(unsafe.Pointer(&cAccountKey.bytes[0]))
	copy((*[40]uint8)(unsafe.Pointer(cArray))[:], accountKey[:])
	return cAccountKey
}

func ConvertToCContractKey(address *types.Address, key []byte) C.CContractKey {
	contractKey := utils.CompositeStorageKey(address, key)
	cContractKey := C.CContractKey{}
	cArray := (*[64]C.uint8_t)(unsafe.Pointer(&cContractKey.bytes[0]))
	copy((*[64]uint8)(unsafe.Pointer(cArray))[:], contractKey[:])
	return cContractKey
}

func ConvertToCCodeKey(address *types.Address, codeHash []byte) C.CCodeKey {
	codeKey := utils.CompositeCodeKey(address, codeHash)
	cCodeKey := C.CCodeKey{}
	cArray := (*[64]C.uint8_t)(unsafe.Pointer(&cCodeKey.bytes[0]))
	copy((*[64]uint8)(unsafe.Pointer(cArray))[:], codeKey[:])
	return cCodeKey
}

func ConvertToCH256(hash common.Hash) C.CH256 {
	ch256 := C.CH256{}
	bytesArray := (*[32]C.uint8_t)(unsafe.Pointer(&ch256.bytes[0]))
	copy((*[32]uint8)(unsafe.Pointer(bytesArray))[:], hash[:])
	return ch256
}

func NewArchiveStateLedger(rep *repo.Repo, height uint64) *ArchiveStateLedger {
	archiveHistoryStateDir := storagemgr.GetLedgerComponentPath(rep, storagemgr.ArchiveHistory)
	archiveJournalDir := storagemgr.GetLedgerComponentPath(rep, storagemgr.ArchiveJournal)

	logDir := path.Join(rep.RepoRoot, "logs")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		panic(err)
	}
	C.set_up_default_logger(C.CString(logDir))

	stateDbPtr := C.new_evm_archive_state_db(C.CString(archiveJournalDir), C.CString(archiveHistoryStateDir))
	return &ArchiveStateLedger{
		repo:           rep,
		logger:         loggers.Logger(loggers.Storage),
		archiveStateDB: stateDbPtr,
		blockHeight:    height,
	}
}

func (r *ArchiveStateLedger) NewView(blockHeader *types.BlockHeader, enableSnapshot bool) (StateLedger, error) {
	return &ArchiveStateLedger{
		repo:           r.repo,
		archiveStateDB: r.archiveStateDB,
		logger:         r.logger,
		blockHeight:    blockHeader.Number,
		rootHash:       blockHeader.StateRoot.ETHHash(),
	}, nil
}

func (r *ArchiveStateLedger) GetOrCreateAccount(address *types.Address) IAccount {
	return r.GetAccount(address)
}

func (r *ArchiveStateLedger) GetAccount(address *types.Address) IAccount {
	r.logger.Infof("[rust-GetAccount] try get account，addr: %v, root hash: %v", address, r.rootHash)

	var length C.uintptr_t
	ptr := C.get_account_at_version(r.archiveStateDB, ConvertToCH256(r.rootHash), ConvertToCAccountKey(address), &length)
	rawAccount := C.GoBytes(unsafe.Pointer(ptr), C.int(length))
	account := &types.InnerAccount{Balance: big.NewInt(0)}
	if err := account.Unmarshal(rawAccount); err != nil {
		panic(err)
	}
	r.logger.Infof("[rust-GetAccount] get from history account trie，addr: %v, account: %v", address, account)
	return NewArchiveAccount(r.archiveStateDB, address, account)
}

func (r *ArchiveStateLedger) Archive(blockHeader *types.BlockHeader, stateJournal *types.StateJournal) error {
	r.logger.Infof("[rust-Archive] archive state journal at height: %v", blockHeader.Number)
	stateJournalBlob := stateJournal.Encode()
	stateJournalPtr := (*C.uchar)(unsafe.Pointer(&stateJournalBlob[0]))
	stateJournalLen := C.uintptr_t(len(stateJournalBlob))
	C.archive(r.archiveStateDB, stateJournalPtr, stateJournalLen)
	return nil
}

func (r *ArchiveStateLedger) GetStateJournal(blockNumber uint64) *types.StateJournal {
	r.logger.Infof("[rust-GetStateJournal] get state journal at height: %v", blockNumber)
	var length C.uintptr_t
	ptr := C.get_state_journal(r.archiveStateDB, C.uint64_t(blockNumber), &length)
	rawStateJournal := C.GoBytes(unsafe.Pointer(ptr), C.int(length))
	res, err := types.DecodeStateJournal(rawStateJournal)
	if err != nil {
		r.logger.Errorf("[rust-GetStateJournal] err:%v", err)
		return nil
	}
	return res
}

func (r *ArchiveStateLedger) ApplyStateJournal(blockNumber uint64, stateJournal *types.StateJournal) error {
	panic("rust archive ledger don't support ApplyStateJournal")
}

func (r *ArchiveStateLedger) GetBalance(address *types.Address) *big.Int {
	account := r.GetOrCreateAccount(address)
	return account.GetBalance()
}

func (r *ArchiveStateLedger) GetState(address *types.Address, bytes []byte) (bool, []byte) {
	account := r.GetOrCreateAccount(address)
	r.logger.Infof("[rust-GetState] get from history storage trie，addr: %v, account: %v", address, account)
	return account.GetState(bytes)
}

func (r *ArchiveStateLedger) GetCode(address *types.Address) []byte {
	account := r.GetOrCreateAccount(address)
	return account.Code()
}

func (r *ArchiveStateLedger) GetNonce(address *types.Address) uint64 {
	account := r.GetOrCreateAccount(address)
	return account.GetNonce()
}

func (r *ArchiveStateLedger) GetCodeHash(address *types.Address) *types.Hash {
	account := r.GetOrCreateAccount(address)
	return types.NewHash(account.CodeHash())
}

func (r *ArchiveStateLedger) GetCodeSize(address *types.Address) int {
	account := r.GetOrCreateAccount(address)
	return len(account.Code())
}

func (r *ArchiveStateLedger) GetCommittedState(address *types.Address, bytes []byte) []byte {
	account := r.GetOrCreateAccount(address)
	return account.GetCommittedState(bytes)
}

func (r *ArchiveStateLedger) SetState(address *types.Address, key []byte, value []byte) {
	account := r.GetOrCreateAccount(address)
	account.SetState(key, value)
}

func (r *ArchiveStateLedger) SetBit256State(address *types.Address, key []byte, value common.Hash) {
	account := r.GetOrCreateAccount(address)
	account.(*ArchiveAccount).SetBit256State(key, value)
}

func (r *ArchiveStateLedger) SetCode(address *types.Address, bytes []byte) {
	account := r.GetOrCreateAccount(address)
	account.SetCodeAndHash(bytes)
}

func (r *ArchiveStateLedger) SetNonce(address *types.Address, u uint64) {
	account := r.GetOrCreateAccount(address)
	account.SetNonce(u)
}

func (r *ArchiveStateLedger) SetBalance(address *types.Address, balance *big.Int) {
	account := r.GetOrCreateAccount(address)
	account.SetBalance(balance)
}

func (r *ArchiveStateLedger) SubBalance(address *types.Address, balance *big.Int) {
	account := r.GetOrCreateAccount(address)
	account.SubBalance(balance)
}

func (r *ArchiveStateLedger) AddBalance(address *types.Address, balance *big.Int) {
	account := r.GetOrCreateAccount(address)
	account.AddBalance(balance)
}

func (r *ArchiveStateLedger) AddRefund(gas uint64) {
	panic("archive state ledger not support AddRefund operation")
}

func (r *ArchiveStateLedger) SubRefund(gas uint64) {
	panic("archive state ledger not support SubRefund operation")
}

func (r *ArchiveStateLedger) GetRefund() uint64 {
	panic("archive state ledger not support GetRefund operation")
}

func (r *ArchiveStateLedger) Commit() (*types.StateJournal, error) {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) SelfDestruct(address *types.Address) bool {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) HasSelfDestructed(address *types.Address) bool {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) Selfdestruct6780(address *types.Address) {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) Exist(address *types.Address) bool {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) Empty(address *types.Address) bool {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) AddressInAccessList(addr types.Address) bool {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) SlotInAccessList(addr types.Address, slot types.Hash) (bool, bool) {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) AddAddressToAccessList(addr types.Address) {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) AddSlotToAccessList(addr types.Address, slot types.Hash) {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) AddPreimage(address types.Hash, preimage []byte) {
	preimagePtr := (*C.uchar)(unsafe.Pointer(&preimage[0]))
	preimageLen := C.uintptr_t(len(preimage))
	C.add_preimage(r.archiveStateDB, ConvertToCH256(address.ETHHash()), preimagePtr, preimageLen)
}

func (r *ArchiveStateLedger) SetTxContext(thash *types.Hash, ti int) {
	panic("archive state ledger not support Commit operation")
}

func (r *ArchiveStateLedger) Clear() {
	// TODO implement me
}

func (r *ArchiveStateLedger) RevertToSnapshot(snapshot int) {
	panic("archive state ledger not support RevertToSnapshot operation")
}

func (r *ArchiveStateLedger) Snapshot() int {
	panic("archive state ledger not support Snapshot operation")
}

func (r *ArchiveStateLedger) setTransientState(addr types.Address, key, value common.Hash) {
	panic("archive state ledger not support setTransientState operation")
}

func (r *ArchiveStateLedger) AddLog(log *types.EvmLog) {
	panic("archive state ledger not support AddLog operation")
}

func (r *ArchiveStateLedger) GetLogs(txHash types.Hash, height uint64) []*types.EvmLog {
	panic("archive state ledger not support GetLogs operation")
}

func (r *ArchiveStateLedger) RollbackState(height uint64, lastStateRoot *types.Hash) error {
	// TODO implement me
	return nil
}

func (r *ArchiveStateLedger) PrepareBlock(lastStateRoot *types.Hash, currentExecutingHeight uint64) {
	panic("archive state ledger not support PrepareBlock operation")
}

func (r *ArchiveStateLedger) ClearChangerAndRefund() {
	panic("archive state ledger not support ClearChangerAndRefund operation")
}

func (r *ArchiveStateLedger) Close() {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) Finalise() {
	panic("archive state ledger not support Finalise operation")
}

func (r *ArchiveStateLedger) Version() uint64 {
	panic("archive state ledger not support Version operation")
}

func (r *ArchiveStateLedger) IterateTrie(snapshotMeta *utils.SnapshotMeta, kv kv.Storage, errC chan error) {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) GetTrieSnapshotMeta() (*utils.SnapshotMeta, error) {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) VerifyTrie(blockHeader *types.BlockHeader) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) Prove(rootHash common.Hash, key []byte) (*jmt.ProofResult, error) {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) GenerateSnapshot(blockHeader *types.BlockHeader, errC chan error) {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) GetHistoryRange() (uint64, uint64) {
	//TODO implement me
	panic("implement me")
}

func (r *ArchiveStateLedger) UpdateChainState(chainState *chainstate.ChainState) {

}

func (r *ArchiveStateLedger) ExportArchivedSnapshot(targetFilePath string) error {
	//TODO implement me
	panic("implement me")
}

var _ StateLedger = (*ArchiveStateLedger)(nil)
