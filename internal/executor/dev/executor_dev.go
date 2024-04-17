package dev

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor"
	"github.com/axiomesh/axiom-ledger/pkg/events"
)

var _ executor.Executor = (*ExecutorDev)(nil)

type ExecutorDev struct {
	blockC             chan *common.CommitEvent
	blockFeed          event.Feed
	blockFeedForRemote event.Feed
	logsFeed           event.Feed
	ctx                context.Context
	cancel             context.CancelFunc

	logger logrus.FieldLogger
}

// New creates executor instance
func New(logger logrus.FieldLogger) (*ExecutorDev, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExecutorDev{
		blockC: make(chan *common.CommitEvent),
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}, nil
}

func (exec *ExecutorDev) Start() error {
	exec.logger.Info("BlockExecutor-DEV started")
	go func() {
		for {
			select {
			case commitEvent := <-exec.blockC:
				exec.processExecuteEvent(commitEvent)
			case <-exec.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (exec *ExecutorDev) processExecuteEvent(commitEvent *common.CommitEvent) {
	current := time.Now()
	block := commitEvent.Block

	txPointerList := make([]*events.TxPointer, len(block.Transactions))
	lo.ForEach(block.Transactions, func(item *types.Transaction, index int) {
		txPointerList[index] = &events.TxPointer{
			Hash:    item.GetHash(),
			Account: item.RbftGetFrom(),
			Nonce:   item.RbftGetNonce(),
		}
	})

	exec.postBlockEvent(block, txPointerList)

	exec.logger.WithFields(logrus.Fields{
		"height": commitEvent.Block.Header.Number,
		"count":  len(commitEvent.Block.Transactions),
		"elapse": time.Since(current),
	}).Info("Executed block")
}

func (exec *ExecutorDev) Stop() error {
	exec.cancel()

	exec.logger.Info("BlockExecutor-DEV stopped")
	return nil
}

func (exec *ExecutorDev) ExecuteBlock(block *common.CommitEvent) {
	exec.processExecuteEvent(block)
}

func (exec *ExecutorDev) AsyncExecuteBlock(commitEvent *common.CommitEvent) {
	exec.blockC <- commitEvent
}

func (exec *ExecutorDev) SubscribeBlockEvent(ch chan<- events.ExecutedEvent) event.Subscription {
	return exec.blockFeed.Subscribe(ch)
}

func (exec *ExecutorDev) SubscribeBlockEventForRemote(ch chan<- events.ExecutedEvent) event.Subscription {
	return exec.blockFeedForRemote.Subscribe(ch)
}

func (exec *ExecutorDev) SubscribeLogsEvent(c chan<- []*types.EvmLog) event.Subscription {
	return exec.logsFeed.Subscribe(c)
}

func (exec *ExecutorDev) NewEvmWithViewLedger(txCtx vm.TxContext, vmConfig vm.Config) (*vm.EVM, error) {
	return nil, nil
}

func (exec *ExecutorDev) postBlockEvent(block *types.Block, txPointerList []*events.TxPointer) {
	exec.blockFeed.Send(events.ExecutedEvent{
		Block:         block,
		TxPointerList: txPointerList,
	})
	exec.blockFeedForRemote.Send(events.ExecutedEvent{
		Block:         block,
		TxPointerList: txPointerList,
	})
}

func (exec *ExecutorDev) GetChainConfig() *params.ChainConfig {
	return nil
}
