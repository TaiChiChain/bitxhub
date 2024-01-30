package ledger

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
)

// ChainLedger handles block, transaction and receipt data.
//
//go:generate mockgen -destination mock_ledger/mock_ledger.go -package mock_ledger -source ledger.go -typed
type ChainLedger interface {
	// GetBlock get block with height
	GetBlock(height uint64) (*types.Block, error)

	// GetBlockSign get the signature of block
	GetBlockSign(height uint64) ([]byte, error)

	// GetBlockByHash get the block using block hash
	GetBlockByHash(hash *types.Hash) (*types.Block, error)

	// GetTransaction get the transaction using transaction hash
	GetTransaction(hash *types.Hash) (*types.Transaction, error)

	// GetTransactionMeta get the transaction meta data
	GetTransactionMeta(hash *types.Hash) (*types.TransactionMeta, error)

	// GetReceipt get the transaction receipt
	GetReceipt(hash *types.Hash) (*types.Receipt, error)

	// GetReceiptsByHeight get the transactions receipts in a block
	GetReceiptsByHeight(height uint64) ([]*types.Receipt, error)

	// PersistExecutionResult persist the execution result
	PersistExecutionResult(block *types.Block, receipts []*types.Receipt) error

	// GetChainMeta get chain meta data
	GetChainMeta() *types.ChainMeta

	// UpdateChainMeta update the chain meta data
	UpdateChainMeta(*types.ChainMeta)

	// LoadChainMeta get chain meta data
	LoadChainMeta() (*types.ChainMeta, error)

	// GetTransactionCount get the transaction count in a block
	GetTransactionCount(height uint64) (uint64, error)

	RollbackBlockChain(height uint64) error

	GetBlockHash(height uint64) *types.Hash

	Close()

	CloseBlockfile()
}

type StateLedger interface {
	StateAccessor

	AddLog(log *types.EvmLog)

	GetLogs(types.Hash, uint64, *types.Hash) []*types.EvmLog

	RollbackState(height uint64, lastStateRoot *types.Hash) error

	PrepareBlock(*types.Hash, *types.Hash, uint64)

	ClearChangerAndRefund()

	// Close release resource
	Close()

	Finalise()

	Version() uint64

	// NewView get a view at specific block. We can enable snapshot if and only if the block were the latest block.
	NewView(block *types.Block, enableSnapshot bool) StateLedger

	// NewViewWithoutCache get a view ledger at specific block. We can enable snapshot if and only if the block were the latest block.
	NewViewWithoutCache(block *types.Block, enableSnapshot bool) StateLedger

	IterateTrie(block *types.Block, kv storage.Storage, errC chan error)

	GetTrieSnapshotMeta() (*SnapshotMeta, error)

	VerifyTrie(block *types.Block) (bool, error)

	Prove(rootHash common.Hash, key []byte) (*jmt.ProofResult, error)

	GenerateSnapshot(block *types.Block, errC chan error)
}

// StateAccessor manipulates the state data
type StateAccessor interface {
	// GetOrCreateAccount
	GetOrCreateAccount(*types.Address) IAccount

	// GetAccount
	GetAccount(*types.Address) IAccount

	// GetBalance
	GetBalance(*types.Address) *big.Int

	// SetBalance
	SetBalance(*types.Address, *big.Int)

	// SubBalance
	SubBalance(*types.Address, *big.Int)

	// AddBalance
	AddBalance(*types.Address, *big.Int)

	// GetState
	GetState(*types.Address, []byte) (bool, []byte)

	// SetState
	SetState(*types.Address, []byte, []byte)

	// SetCode
	SetCode(*types.Address, []byte)

	// GetCode
	GetCode(*types.Address) []byte

	// SetNonce
	SetNonce(*types.Address, uint64)

	// GetNonce
	GetNonce(*types.Address) uint64

	// GetCodeHash
	GetCodeHash(*types.Address) *types.Hash

	// GetCodeSize
	GetCodeSize(*types.Address) int

	// AddRefund
	AddRefund(uint64)

	// SubRefund
	SubRefund(uint64)

	// GetRefund
	GetRefund() uint64

	// GetCommittedState
	GetCommittedState(*types.Address, []byte) []byte

	// Commit commits the state data
	Commit() (*types.Hash, error)

	// Suicide
	Suicide(*types.Address) bool

	// HasSuicide
	HasSuicide(*types.Address) bool

	// Exist
	Exist(*types.Address) bool

	// Empty
	Empty(*types.Address) bool

	// AddressInAccessList
	AddressInAccessList(types.Address) bool

	// SlotInAccessList
	SlotInAccessList(types.Address, types.Hash) (bool, bool)

	// AddAddressToAccessList
	AddAddressToAccessList(types.Address)

	// AddSlotToAccessList
	AddSlotToAccessList(types.Address, types.Hash)

	// AddPreimage
	AddPreimage(types.Hash, []byte)

	// Set tx context for state db
	SetTxContext(thash *types.Hash, txIndex int)

	// Clear
	Clear()

	// RevertToSnapshot
	RevertToSnapshot(int)

	// Snapshot
	Snapshot() int
}

type IAccount interface {
	fmt.Stringer

	GetAddress() *types.Address

	GetState(key []byte) (bool, []byte)

	GetCommittedState(key []byte) []byte

	SetState(key []byte, value []byte)

	SetCodeAndHash(code []byte)

	Code() []byte

	CodeHash() []byte

	SetNonce(nonce uint64)

	GetNonce() uint64

	GetBalance() *big.Int

	SetBalance(balance *big.Int)

	SubBalance(amount *big.Int)

	AddBalance(amount *big.Int)

	Finalise() [][]byte

	IsEmpty() bool

	Suicided() bool

	SetSuicided(bool)

	SetEnableExpensiveMetric(bool)

	GetStorageRootHash() common.Hash

	GetStorageRoot() common.Hash
}

var _ vm.StateDB = (*EvmStateDBAdaptor)(nil)

// EvmStateDBAdaptor wraps StateLedger with Wrapper mode
type EvmStateDBAdaptor struct {
	StateLedger StateLedger
}
