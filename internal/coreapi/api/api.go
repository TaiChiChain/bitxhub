package api

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	vm "github.com/axiomesh/eth-kit/evm"
)

//go:generate mockgen -destination mock_api/mock_api.go -package mock_api -source api.go -typed
type CoreAPI interface {
	Broker() BrokerAPI
	Chain() ChainAPI
	Feed() FeedAPI
	Gas() GasAPI
	TxPool() TxPoolAPI
}

type BrokerAPI interface {
	HandleTransaction(tx *types.Transaction) error
	GetTransaction(*types.Hash) (*types.Transaction, error)
	GetTransactionMeta(*types.Hash) (*types.TransactionMeta, error)
	GetReceipt(*types.Hash) (*types.Receipt, error)
	GetBlock(mode string, key string) (*types.Block, error)
	GetBlocks(start uint64, end uint64) ([]*types.Block, error)
	GetViewStateLedger() ledger.StateLedger
	GetEvm(mes *vm.Message, vmConfig *vm.Config) (*vm.EVM, error)
	GetNativeVm() common.VirtualMachine
	ConsensusReady() error
	GetBlockHeaders(start uint64, end uint64) ([]*types.BlockHeader, error)

	ChainConfig() *params.ChainConfig
	StateAtTransaction(block *types.Block, txIndex int, reexec uint64) (*vm.Message, vm.BlockContext, *ledger.StateLedger, error)
}

type NetworkAPI interface {
	PeerInfo() ([]byte, error)
}

type ChainAPI interface {
	Status() string
	Meta() (*types.ChainMeta, error)
	TPS(begin, end uint64) (uint64, error)
}

type FeedAPI interface {
	SubscribeLogsEvent(chan<- []*types.EvmLog) event.Subscription
	SubscribeNewTxEvent(chan<- []*types.Transaction) event.Subscription
	SubscribeNewBlockEvent(chan<- events.ExecutedEvent) event.Subscription
	BloomStatus() (uint64, uint64)
}

type GasAPI interface {
	GetGasPrice() (uint64, error)
	GetCurrentGasPrice(blockHeight uint64) (uint64, error)
}

type TxPoolAPI interface {
	GetPendingTxCountByAccount(account string) uint64
	GetTotalPendingTxCount() uint64
	GetTransaction(hash *types.Hash) *types.Transaction
	GetAccountMeta(account string, full bool) any
	GetMeta(full bool) any
	GetChainInfo() any
}
