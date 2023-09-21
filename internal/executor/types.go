package executor

import (
	"github.com/ethereum/go-ethereum/event"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/order/common"
	"github.com/axiomesh/axiom-ledger/pkg/model/events"
	vm "github.com/axiomesh/eth-kit/evm"
)

type Executor interface {
	// Start
	Start() error

	// Stop
	Stop() error

	// ExecutorBlock
	ExecuteBlock(commitEvent *common.CommitEvent)

	// ApplyReadonlyTransactions execute readonly tx
	ApplyReadonlyTransactions(txs []*types.Transaction) []*types.Receipt

	// SubscribeBlockEvent
	SubscribeBlockEvent(chan<- events.ExecutedEvent) event.Subscription

	// SubscribeBlockEventForRemote
	SubscribeBlockEventForRemote(chan<- events.ExecutedEvent) event.Subscription

	// SubscribeLogEvent
	SubscribeLogsEvent(chan<- []*types.EvmLog) event.Subscription

	NewEvmWithViewLedger(txCtx vm.TxContext, vmConfig vm.Config) (*vm.EVM, error)
}
