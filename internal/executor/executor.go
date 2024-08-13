package executor

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	blockChanNumber = 1024
)

var _ Executor = (*BlockExecutor)(nil)

type AfterBlockHook struct {
	Name string
	Func func(block *types.Block) error
}

// BlockExecutor executes block from consensus
type BlockExecutor struct {
	ledger             *ledger.Ledger
	logger             logrus.FieldLogger
	blockC             chan *common.CommitEvent
	cumulativeGasUsed  uint64
	currentHeight      uint64
	currentBlockHash   *types.Hash
	blockFeed          event.Feed
	blockFeedForRemote event.Feed
	logsFeed           event.Feed
	ctx                context.Context
	cancel             context.CancelFunc

	evm         *vm.EVM
	evmChainCfg *params.ChainConfig
	gasLimit    uint64
	rep         *repo.Repo
	chainState  *chainstate.ChainState

	nvm             syscommon.VirtualMachine
	afterBlockHooks []AfterBlockHook

	enableParallelExecute bool

	stateDbPool *StateDbPool
}

// New creates executor instance
func New(rep *repo.Repo, ledger *ledger.Ledger, chainState *chainstate.ChainState) (*BlockExecutor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	blockExecutor := &BlockExecutor{
		ledger:                ledger,
		logger:                loggers.Logger(loggers.Executor),
		ctx:                   ctx,
		chainState:            chainState,
		cancel:                cancel,
		blockC:                make(chan *common.CommitEvent, blockChanNumber),
		cumulativeGasUsed:     0,
		currentHeight:         ledger.ChainLedger.GetChainMeta().Height,
		currentBlockHash:      ledger.ChainLedger.GetChainMeta().BlockHash,
		evmChainCfg:           newEVMChainCfg(rep.GenesisConfig),
		rep:                   rep,
		gasLimit:              rep.GenesisConfig.EpochInfo.FinanceParams.GasLimit,
		enableParallelExecute: rep.Config.EnableParallel,
		stateDbPool:           NewStateDbPool(),
	}

	blockExecutor.evm = newEvm(1, uint64(0), blockExecutor.evmChainCfg, blockExecutor.ledger.StateLedger, blockExecutor.ledger.ChainLedger, "")

	// initialize native vm
	blockExecutor.nvm = system.New()
	blockExecutor.afterBlockHooks = []AfterBlockHook{
		{
			Name: "recordReward",
			Func: blockExecutor.recordReward,
		},
		{
			Name: "tryTurnIntoNewEpoch",
			Func: blockExecutor.tryTurnIntoNewEpoch,
		},
	}

	return blockExecutor, nil
}

// Start starts executor
func (exec *BlockExecutor) Start() error {
	go exec.listenExecuteEvent()

	exec.logger.WithFields(logrus.Fields{
		"height": exec.currentHeight,
		"hash":   exec.currentBlockHash.String(),
	}).Infof("BlockExecutor started")

	return nil
}

// Stop stops executor
func (exec *BlockExecutor) Stop() error {
	exec.cancel()

	exec.logger.Info("BlockExecutor stopped")

	return nil
}

// ExecuteBlock executes block from consensus
func (exec *BlockExecutor) ExecuteBlock(block *common.CommitEvent) {
	if exec.enableParallelExecute {
		exec.processExecuteEventParallel(block)
	} else {
		exec.processExecuteEvent(block)
	}

}

func (exec *BlockExecutor) AsyncExecuteBlock(block *common.CommitEvent) {
	exec.blockC <- block
}

func (exec *BlockExecutor) CurrentHeader() *types.BlockHeader {
	header, _ := exec.ledger.ChainLedger.GetBlockHeader(exec.currentHeight)
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println(exec.currentHeight)
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	exec.logger.Println("----------------------------------------------")
	return header
}

// SubscribeBlockEvent registers a subscription of ExecutedEvent.
func (exec *BlockExecutor) SubscribeBlockEvent(ch chan<- events.ExecutedEvent) event.Subscription {
	return exec.blockFeed.Subscribe(ch)
}

// SubscribeBlockEventForRemote registers a subscription of ExecutedEvent.
func (exec *BlockExecutor) SubscribeBlockEventForRemote(ch chan<- events.ExecutedEvent) event.Subscription {
	return exec.blockFeedForRemote.Subscribe(ch)
}

func (exec *BlockExecutor) SubscribeLogsEvent(ch chan<- []*types.EvmLog) event.Subscription {
	return exec.logsFeed.Subscribe(ch)
}

func (exec *BlockExecutor) listenExecuteEvent() {
	for {
		select {
		case <-exec.ctx.Done():
			close(exec.blockC)
			return
		case commitEvent := <-exec.blockC:
			if exec.enableParallelExecute {
				exec.processExecuteEventParallel(commitEvent)
			} else {
				exec.processExecuteEvent(commitEvent)
			}
		}
	}
}

func newEVMChainCfg(genesisConfig *repo.GenesisConfig) *params.ChainConfig {
	shanghaiTime := uint64(0)
	CancunTime := uint64(0)
	PragueTime := uint64(0)

	return &params.ChainConfig{
		ChainID:                 big.NewInt(int64(genesisConfig.ChainID)),
		HomesteadBlock:          big.NewInt(0),
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(0),
		ShanghaiTime:            &shanghaiTime,
		CancunTime:              &CancunTime,
		PragueTime:              &PragueTime,
	}
}
