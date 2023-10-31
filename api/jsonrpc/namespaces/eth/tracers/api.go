package tracers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/eth/tracers/logger"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

const (
	// defaultTraceTimeout is the amount of time a single transaction can execute
	// by default before being forcefully aborted.
	defaultTraceTimeout = 5 * time.Second

	// defaultTraceReexec is the number of blocks the tracer is willing to go api.Broker()
	// and reexecute to produce missing historical state necessary to run a specific
	// trace.
	defaultTraceReexec = uint64(128)

	// defaultTracechainMemLimit is the size of the triedb, at which traceChain
	// switches over and tries to use a disk-api.Broker()ed database instead of building
	// on top of memory.
	// For non-archive nodes, this limit _will_ be overblown, as disk-api.Broker()ed tries
	// will only be found every ~15K blocks or so.
	defaultTracechainMemLimit = common.StorageSize(500 * 1024 * 1024)

	// maximumPendingTraceStates is the maximum number of states allowed waiting
	// for tracing. The creation of trace state will be paused if the unused
	// trace states exceed this limit.
	maximumPendingTraceStates = 128
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

var errTxNotFound = errors.New("transaction not found")

func (api *TracerAPI) TraceTransaction(hash common.Hash, config *TraceConfig) (interface{}, error) {
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
	block, err := api.api.Broker().GetBlock("HEIGHT", fmt.Sprintf("%d", meta.BlockHeight))
	if err != nil {
		return nil, err
	}
	msg, vmctx, statedb, err := api.api.Broker().StateAtTransaction(block, int(meta.Index), reexec)
	if err != nil {
		return nil, err
	}
	txctx := &Context{
		BlockHash:   block.BlockHash.ETHHash(),
		BlockNumber: new(big.Int).SetUint64(block.Height()),
		TxIndex:     int(meta.Index),
		TxHash:      hash,
	}

	return api.traceTx(msg, txctx, vmctx, *statedb, config)

}

func (api *TracerAPI) traceTx(message *vm.Message, txctx *Context, vmctx vm.BlockContext, statedb ledger.StateLedger, config *TraceConfig) (interface{}, error) {
	var (
		tracer    Tracer
		err       error
		timeout   = defaultTraceTimeout
		txContext = vm.NewEVMTxContext(message)
	)
	if config == nil {
		config = &TraceConfig{}
	}
	// Default tracer is the struct logger
	tracer = logger.NewStructLogger(config.Config)
	if config.Tracer != nil {
		tracer, err = DefaultDirectory.New(*config.Tracer, txctx, config.TracerConfig)
		if err != nil {
			return nil, err
		}
	}
	vmenv := vm.NewEVM(vmctx, txContext, statedb, api.api.Broker().ChainConfig(), vm.Config{Tracer: tracer, NoBaseFee: true})
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

	// Call Prepare to clear out the statedb access list
	statedb.SetTxContext(types.NewHash(txctx.BlockHash.Bytes()), txctx.TxIndex)
	if _, err = vm.ApplyMessage(vmenv, message, new(vm.GasPool).AddGas(message.GasLimit)); err != nil {
		return nil, fmt.Errorf("tracing failed: %w", err)
	}
	return tracer.GetResult()

}
