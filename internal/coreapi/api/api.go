/*
 * @Date: 2024-05-27 09:49:49
 * @LastEditors: levi9311 790890362@qq.com
 * @LastEditTime: 2024-05-27 14:58:38
 * @FilePath: /axiom-ledger/internal/coreapi/api/api.go
 */
package api

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/events"
)

//go:generate mockgen -destination mock_api/mock_api.go -package mock_api -source api.go -typed
type CoreAPI interface {
	Broker() BrokerAPI
	Chain() ChainAPI
	Feed() FeedAPI
	TxPool() TxPoolAPI
	ChainState() *chainstate.ChainState
}

type BrokerAPI interface {
	HandleTransaction(tx *types.Transaction) error
	GetTransaction(*types.Hash) (*types.Transaction, error)
	GetTransactionMeta(*types.Hash) (*types.TransactionMeta, error)
	GetReceipt(*types.Hash) (*types.Receipt, error)
	GetReceipts(blockNum uint64) ([]*types.Receipt, error)
	GetViewStateLedger() ledger.StateLedger
	GetEvm(mes *core.Message, vmConfig *vm.Config) (*vm.EVM, error)
	ConsensusReady() error

	ChainConfig() *params.ChainConfig
	StateAtTransaction(block *types.Block, txIndex int, reexec uint64) (*core.Message, vm.BlockContext, *ledger.StateLedger, error)

	GetBlockHeaderByNumber(height uint64) (*types.BlockHeader, error)
	GetBlockHeaderByHash(hash *types.Hash) (*types.BlockHeader, error)

	GetBlockExtra(height uint64) (*types.BlockExtra, error)

	GetBlockTxHashList(height uint64) ([]*types.Hash, error)
	GetBlockTxList(height uint64) ([]*types.Transaction, error)

	GetSyncProgress() *common.SyncProgress
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
	ServiceFilter(session *bloombits.MatcherSession)
}

type TxPoolAPI interface {
	GetPendingTxCountByAccount(account string) uint64
	GetTotalPendingTxCount() uint64
	GetTransaction(hash *types.Hash) *types.Transaction
	GetAccountMeta(account string, full bool) any
	GetMeta(full bool) any
	GetChainInfo() any
	IsStarted() bool
}
