package tracers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/executor"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	// defaultTraceTimeout is the amount of time a single transaction can execute
	// by default before being forcefully aborted.
	defaultTraceTimeout = 5 * time.Second

	// defaultTraceReexec is the number of blocks the tracer is willing to go api.Broker()
	// and reexecute to produce missing historical state necessary to run a specific
	// trace.
	defaultTraceReexec = uint64(128)
)

type StateReleaseFunc func()

type TracerAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

// todo
func NewTracerAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *TracerAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &TracerAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

type TraceConfig struct {
	*logger.Config
	Tracer  *string
	Timeout *string
	Reexec  *uint64

	// Config specific to given tracer. Note struct logger
	// config are historically embedded in main object.
	TracerConfig json.RawMessage
}

// TraceCallConfig is the config for traceCall API. It holds one more
// field to override the state for tracing.
type TraceCallConfig struct {
	TraceConfig
	StateOverrides *StateOverride
	BlockOverrides *BlockOverrides
}

var errTxNotFound = errors.New("transaction not found")

func (api *TracerAPI) TraceTransaction(hash common.Hash, config *TraceConfig) (any, error) {
	txHash := types.NewHash(hash.Bytes())
	tx, err := api.api.Broker().GetTransaction(txHash)
	if err != nil {
		return nil, nil
	}
	if tx == nil {
		return nil, errTxNotFound
	}

	meta, err := api.api.Broker().GetTransactionMeta(txHash)
	if err != nil {
		return nil, fmt.Errorf("get tx meta from ledger: %w", err)
	}

	if meta.BlockHeight == 1 {
		return nil, errors.New("genesis is not traceable")
	}

	reexec := defaultTraceReexec
	if config != nil && config.Reexec != nil {
		reexec = *config.Reexec
	}
	blockHeader, err := api.api.Broker().GetBlockHeaderByNumber(meta.BlockHeight)
	if err != nil {
		return nil, err
	}
	blockTxList, err := api.api.Broker().GetBlockTxList(blockHeader.Number)
	if err != nil {
		return nil, err
	}

	block := &types.Block{
		Header:       blockHeader,
		Transactions: blockTxList,
	}
	msg, vmctx, statedb, err := api.api.Broker().StateAtTransaction(block, int(meta.Index), reexec)
	if err != nil {
		return nil, err
	}
	txctx := &tracers.Context{
		BlockHash:   block.Hash().ETHHash(),
		BlockNumber: new(big.Int).SetUint64(block.Height()),
		TxIndex:     int(meta.Index),
		TxHash:      hash,
	}

	return api.traceTx(msg, txctx, vmctx, *statedb, config)
}

func (api *TracerAPI) traceTx(message *core.Message, txctx *tracers.Context, vmctx vm.BlockContext, statedb ledger.StateLedger, config *TraceConfig) (any, error) {
	var (
		tracer    tracers.Tracer
		err       error
		timeout   = defaultTraceTimeout
		txContext = core.NewEVMTxContext(message)
	)
	if config == nil {
		config = &TraceConfig{}
	}
	// Default tracer is the struct logger
	tracer = logger.NewStructLogger(config.Config)
	if config.Tracer != nil {
		tracer, err = tracers.DefaultDirectory.New(*config.Tracer, txctx, config.TracerConfig)
		if err != nil {
			return nil, err
		}
	}
	vmenv := vm.NewEVM(vmctx, txContext, &ledger.EvmStateDBAdaptor{StateLedger: statedb}, api.api.Broker().ChainConfig(), vm.Config{Tracer: tracer, NoBaseFee: true})
	// Define a meaningful timeout of a single transaction trace
	if config.Timeout != nil {
		if timeout, err = time.ParseDuration(*config.Timeout); err != nil {
			return nil, err
		}
	}
	deadlineCtx, cancel := context.WithTimeout(api.ctx, timeout)
	go func() {
		<-deadlineCtx.Done()
		if errors.Is(deadlineCtx.Err(), context.DeadlineExceeded) {
			tracer.Stop(errors.New("execution timeout"))
			// Stop evm execution. Note cancellation is not necessarily immediate.
			vmenv.Cancel()
		}
	}()
	defer cancel()

	// CrossCallEVMContract Prepare to clear out the statedb access list
	statedb.SetTxContext(types.NewHash(txctx.BlockHash.Bytes()), txctx.TxIndex)
	if _, err = core.ApplyMessage(vmenv, message, new(core.GasPool).AddGas(message.GasLimit)); err != nil {
		api.logger.Errorf("trace failed: %w", err)
		return nil, fmt.Errorf("tracing failed: %w", err)
	}
	traceRes, err := tracer.GetResult()
	api.logger.Debugf("trace call, result: %+v, %s", traceRes, err)
	return traceRes, err
}

func (api *TracerAPI) TraceCall(args types.CallArgs, blockNrOrHash *rpctypes.BlockNumberOrHash, config *TraceCallConfig) (any, error) {
	api.logger.Debugf("trace call, args: %+v, block number or hash: %+v, config: %+v", args, blockNrOrHash, config)
	// Try to retrieve the specified block
	var (
		err         error
		blockHeader *types.BlockHeader
	)
	if hash, ok := blockNrOrHash.Hash(); ok {
		blockHeader, err = api.api.Broker().GetBlockHeaderByHash(types.NewHash(hash.Bytes()))
	} else if blockNum, ok := blockNrOrHash.Number(); ok {
		if blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber {
			meta, _ := api.api.Chain().Meta()
			blockNum = rpctypes.BlockNumber(meta.Height)
		}
		blockHeader, err = api.api.Broker().GetBlockHeaderByNumber(uint64(blockNum))
	} else {
		return nil, errors.New("invalid arguments; neither block nor hash specified")
	}
	if err != nil {
		return nil, err
	}

	statedb, err := api.api.Broker().GetViewStateLedger().NewView(blockHeader, false)
	if err != nil {
		return nil, err
	}

	vmctx := executor.NewEVMBlockContextAdaptor(blockHeader.Number, uint64(blockHeader.Timestamp), syscommon.StakingManagerContractAddr, nil)
	// Apply the customization rules if required.
	if config != nil {
		if err := config.StateOverrides.Apply(statedb); err != nil {
			return nil, err
		}
		config.BlockOverrides.Apply(&vmctx)
	}
	// Execute the trace
	msg, err := executor.CallArgsToMessage(&args, api.rep.Config.JsonRPC.GasCap, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	var traceConfig *TraceConfig
	if config != nil {
		traceConfig = &config.TraceConfig
	}
	return api.traceTx(msg, new(tracers.Context), vmctx, statedb, traceConfig)
}

// OverrideAccount indicates the overriding fields of account during the execution
// of a message call.
// Note, state and stateDiff can't be specified at the same time. If state is
// set, message execution will only use the data in the given state. Otherwise
// if statDiff is set, all diff will be applied first and then execute the call
// message.
type OverrideAccount struct {
	Nonce     *hexutil.Uint64              `json:"nonce"`
	Code      *hexutil.Bytes               `json:"code"`
	Balance   **hexutil.Big                `json:"balance"`
	State     *map[common.Hash]common.Hash `json:"state"`
	StateDiff *map[common.Hash]common.Hash `json:"stateDiff"`
}

// StateOverride is the collection of overridden accounts.
type StateOverride map[common.Address]OverrideAccount

// Apply overrides the fields of specified accounts into the given state.
func (diff *StateOverride) Apply(state ledger.StateLedger) error {
	if diff == nil {
		return nil
	}
	for addr, account := range *diff {
		taddr := types.NewAddressByStr(addr.String())
		// Override account nonce.
		if account.Nonce != nil {
			state.SetNonce(taddr, uint64(*account.Nonce))
		}
		// Override account(contract) code.
		if account.Code != nil {
			state.SetCode(taddr, *account.Code)
		}
		// Override account balance.
		if account.Balance != nil {
			state.SetBalance(taddr, (*big.Int)(*account.Balance))
		}
		if account.State != nil && account.StateDiff != nil {
			return fmt.Errorf("account %s has both 'state' and 'stateDiff'", addr.String())
		}
		// Replace entire state if caller requires.
		if account.State != nil {
			acc := state.GetAccount(taddr)
			for k, v := range *account.State {
				acc.SetState(k.Bytes(), v.Bytes())
			}
		}
		// Apply state diff into specified accounts.
		if account.StateDiff != nil {
			for key, value := range *account.StateDiff {
				state.SetState(taddr, key.Bytes(), value.Bytes())
			}
		}
	}
	// Now finalize the changes. Finalize is normally performed between transactions.
	// By using finalize, the overrides are semantically behaving as
	// if they were created in a transaction just before the tracing occur.
	state.Finalise()
	return nil
}

// BlockOverrides is a set of header fields to override.
type BlockOverrides struct {
	Number     *hexutil.Big
	Difficulty *hexutil.Big
	Time       *hexutil.Uint64
	GasLimit   *hexutil.Uint64
	Coinbase   *common.Address
	Random     *common.Hash
	BaseFee    *hexutil.Big
}

// Apply overrides the given header fields into the given block context.
func (diff *BlockOverrides) Apply(blockCtx *vm.BlockContext) {
	if diff == nil {
		return
	}
	if diff.Number != nil {
		blockCtx.BlockNumber = diff.Number.ToInt()
	}
	if diff.Difficulty != nil {
		blockCtx.Difficulty = diff.Difficulty.ToInt()
	}
	if diff.Time != nil {
		blockCtx.Time = uint64(*diff.Time)
	}
	if diff.GasLimit != nil {
		blockCtx.GasLimit = uint64(*diff.GasLimit)
	}
	if diff.Coinbase != nil {
		blockCtx.Coinbase = *diff.Coinbase
	}
	if diff.Random != nil {
		blockCtx.Random = diff.Random
	}
	if diff.BaseFee != nil {
		blockCtx.BaseFee = diff.BaseFee.ToInt()
	}
}
