package eth

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethhexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/axiomesh/eth-kit/adaptor"
	vm "github.com/axiomesh/eth-kit/evm"
)

// BlockChain API provides an API for accessing blockchain data
type BlockChainAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

func NewBlockChainAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *BlockChainAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockChainAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

// ChainId returns the chain's identifier in hex format
func (api *BlockChainAPI) ChainId() (ret ethhexutil.Uint, err error) { // nolint
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debug("eth_chainId")
	return ethhexutil.Uint(api.rep.GenesisConfig.ChainID), nil
}

// BlockNumber returns the current block number.
func (api *BlockChainAPI) BlockNumber() (ret ethhexutil.Uint64, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debug("eth_blockNumber")
	meta, err := api.api.Chain().Meta()
	if err != nil {
		return 0, err
	}

	return ethhexutil.Uint64(meta.Height), nil
}

// GetBalance returns the provided account's balance, blockNum is ignored.
func (api *BlockChainAPI) GetBalance(address common.Address, blockNrOrHash *rpctypes.BlockNumberOrHash) (ret *ethhexutil.Big, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getBalance, address: %s, block number : %d", address.String())

	stateLedger, err := getStateLedgerAt(api.api, blockNrOrHash)
	if err != nil {
		return nil, err
	}

	balance := stateLedger.GetBalance(types.NewAddress(address.Bytes()))
	api.logger.Debugf("balance: %d", balance)

	return (*ethhexutil.Big)(balance), nil
}

type AccountResult struct {
	Address      common.Address    `json:"address"`
	AccountProof []string          `json:"accountProof"`
	Balance      *ethhexutil.Big   `json:"balance"`
	CodeHash     common.Hash       `json:"codeHash"`
	Nonce        ethhexutil.Uint64 `json:"nonce"`
	StorageHash  common.Hash       `json:"storageHash"`
	StorageProof []StorageResult   `json:"storageProof"`
}

type StorageResult struct {
	Key   string          `json:"key"`
	Value *ethhexutil.Big `json:"value"`
	Proof []string        `json:"proof"`
}

// todo
// GetProof returns the Merkle-proof for a given account and optionally some storage keys.
func (api *BlockChainAPI) GetProof(address common.Address, storageKeys []string, blockNrOrHash *rpctypes.BlockNumberOrHash) (ret *AccountResult, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	return nil, ErrNotSupportApiError
}

// GetBlockByNumber returns the block identified by number.
func (api *BlockChainAPI) GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (ret map[string]any, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getBlockByNumber, number: %d, full: %v", blockNum, fullTx)

	if blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber {
		meta, err := api.api.Chain().Meta()
		if err != nil {
			return nil, err
		}
		blockNum = rpctypes.BlockNumber(meta.Height)
	}

	block, err := api.api.Broker().GetBlock("HEIGHT", fmt.Sprintf("%d", blockNum))
	if err != nil {
		if errors.Is(err, ledger.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return formatBlock(api.api, api.rep.GenesisConfig, block, fullTx)
}

// GetBlockByHash returns the block identified by hash.
func (api *BlockChainAPI) GetBlockByHash(hash common.Hash, fullTx bool) (ret map[string]any, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getBlockByHash, hash: %s, full: %v", hash.String(), fullTx)

	block, err := api.api.Broker().GetBlock("HASH", hash.String())
	if err != nil {
		if errors.Is(err, ledger.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return formatBlock(api.api, api.rep.GenesisConfig, block, fullTx)
}

// GetCode returns the contract code at the given address, blockNum is ignored.
func (api *BlockChainAPI) GetCode(address common.Address, blockNrOrHash *rpctypes.BlockNumberOrHash) (ret ethhexutil.Bytes, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getCode, address: %s", address.String())

	stateLedger, err := getStateLedgerAt(api.api, blockNrOrHash)
	if err != nil {
		return nil, err
	}

	code := stateLedger.GetCode(types.NewAddress(address.Bytes()))

	return code, nil
}

// GetStorageAt returns the contract storage at the given address and key, blockNum is ignored.
func (api *BlockChainAPI) GetStorageAt(address common.Address, key string, blockNrOrHash *rpctypes.BlockNumberOrHash) (ret ethhexutil.Bytes, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getStorageAt, address: %s, key: %s", address, key)

	stateLedger, err := getStateLedgerAt(api.api, blockNrOrHash)
	if err != nil {
		return nil, err
	}

	hash, err := hexutil.DecodeHash(key)
	if err != nil {
		return nil, err
	}

	ok, val := stateLedger.GetState(types.NewAddress(address.Bytes()), hash.Bytes())
	if !ok {
		return nil, nil
	}

	return val, nil
}

// Call performs a raw contract call.
func (api *BlockChainAPI) Call(args types.CallArgs, blockNrOrHash *rpctypes.BlockNumberOrHash, _ *map[common.Address]rpctypes.Account) (ret ethhexutil.Bytes, err error) {
	defer func(start time.Time) {
		invokeCallContractDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_call, args: %v", args)

	receipt, err := DoCall(api.ctx, blockNrOrHash, api.rep.Config.Executor.EVM, api.api, args, api.rep.Config.JsonRPC.EVMTimeout.ToDuration(), api.rep.Config.JsonRPC.GasCap, api.logger)
	if err != nil {
		return nil, err
	}
	api.logger.Debugf("receipt: %v", receipt)
	if len(receipt.Revert()) > 0 {
		return nil, newRevertError(receipt.Revert())
	}

	return receipt.Return(), receipt.Err
}

// DoCall todo call with historical ledger
func DoCall(ctx context.Context, blockNrOrHash *rpctypes.BlockNumberOrHash, evmCfg repo.EVM, api api.CoreAPI, args types.CallArgs, timeout time.Duration, globalGasCap uint64, logger logrus.FieldLogger) (*vm.ExecutionResult, error) {
	defer func(start time.Time) { logger.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// GET EVM Instance
	msg, err := adaptor.CallArgsToMessage(&args, globalGasCap, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	// use copy state ledger to call
	stateLedger, err := getStateLedgerAt(api, blockNrOrHash)
	if err != nil {
		return nil, err
	}
	stateLedger.SetTxContext(types.NewHash([]byte("mockTx")), 0)

	// check if call system contract
	nvm := api.Broker().GetNativeVm()
	if msg.To != nil && nvm.IsSystemContract(types.NewAddress(msg.To.Bytes())) {
		chainMeta, err := api.Chain().Meta()
		if err != nil {
			return nil, err
		}
		nvm.Reset(chainMeta.Height, stateLedger)
		return nvm.Run(msg)
	}

	evm, err := api.Broker().GetEvm(msg, &vm.Config{NoBaseFee: true, DisableMaxCodeSizeLimit: evmCfg.DisableMaxCodeSizeLimit})
	if err != nil {
		return nil, errors.New("error get evm")
	}

	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	txContext := vm.NewEVMTxContext(msg)
	evm.Reset(txContext, stateLedger)
	gp := new(vm.GasPool).AddGas(math.MaxUint64)
	result, err := vm.ApplyMessage(evm, msg, gp)

	// If the timer caused an abort, return an appropriate error message
	if evm.Cancelled() {
		return nil, fmt.Errorf("execution aborted (timeout = %v)", timeout)
	}
	if err != nil {
		// logger.Errorf("err: %w (supplied gas %d)", err, msg.GasLimit)
		return result, err
	}

	return result, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
// It adds 2,000 gas to the returned value instead of using the gas adjustment
// param from the SDK.
func (api *BlockChainAPI) EstimateGas(args types.CallArgs, blockNrOrHash *rpctypes.BlockNumberOrHash) (ret ethhexutil.Uint64, err error) {
	defer func(start time.Time) {
		invokeCallContractDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_estimateGas, args: %s", args)

	// Judge whether this is system contract
	nvm := api.api.Broker().GetNativeVm()
	if args.To != nil && nvm.IsSystemContract(types.NewAddress(args.To.Bytes())) {
		gas, err := nvm.EstimateGas(&args)
		return ethhexutil.Uint64(gas), err
	}

	// Determine the highest gas limit can be used during the estimation.
	// if args.Gas == nil || uint64(*args.Gas) < params.TxGas {
	// 	// Retrieve the block to act as the gas ceiling
	// 	args.Gas = (*ethhexutil.Uint64)(&api.rep.GasLimit)
	// }
	// Determine the lowest and highest possible gas limits to binary search in between
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)
	if args.Gas != nil && uint64(*args.Gas) >= params.TxGas {
		hi = uint64(*args.Gas)
	} else {
		// todo : api need get current epochInfo
		hi = api.rep.GenesisConfig.EpochInfo.FinanceParams.GasLimit
	}

	var feeCap *big.Int
	if args.GasPrice != nil && (args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil) {
		return 0, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	} else if args.GasPrice != nil {
		feeCap = args.GasPrice.ToInt()
	} else if args.MaxFeePerGas != nil {
		feeCap = args.MaxFeePerGas.ToInt()
	} else {
		feeCap = common.Big0
	}
	if feeCap.BitLen() != 0 {
		stateLedger, err := getStateLedgerAt(api.api, blockNrOrHash)
		if err != nil {
			return 0, err
		}
		balance := stateLedger.GetBalance(types.NewAddress(args.From.Bytes()))
		api.logger.Debugf("balance: %d", balance)
		available := new(big.Int).Set(balance)
		if args.Value != nil {
			if args.Value.ToInt().Cmp(available) >= 0 {
				return 0, core.ErrInsufficientFundsForTransfer
			}
			available.Sub(available, args.Value.ToInt())
		}
		allowance := new(big.Int).Div(available, feeCap)
		if allowance.IsUint64() && hi > allowance.Uint64() {
			transfer := args.Value
			if transfer == nil {
				transfer = new(ethhexutil.Big)
			}
			api.logger.Warn("Gas estimation capped by limited funds", "original", hi, "balance", balance,
				"sent", transfer.ToInt(), "maxFeePerGas", feeCap, "fundable", allowance)
			hi = allowance.Uint64()
		}
	}

	gasCap := api.rep.Config.JsonRPC.GasCap
	if gasCap != 0 && hi > gasCap {
		api.logger.Warn("Caller gas above allowance, capping", "requested", hi, "cap", gasCap)
		hi = gasCap
	}

	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *vm.ExecutionResult, error) {
		args.Gas = (*ethhexutil.Uint64)(&gas)

		result, err := DoCall(api.ctx, blockNrOrHash, api.rep.Config.Executor.EVM, api.api, args, api.rep.Config.JsonRPC.EVMTimeout.ToDuration(), api.rep.Config.JsonRPC.GasCap, api.logger)
		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, err
		}
		return result.Failed(), result, nil
	}

	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		failed, _, err := executable(mid)
		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		failed, ret, err := executable(hi)
		if err != nil {
			return 0, err
		}
		if failed {
			if ret != nil && ret.Err != vm.ErrOutOfGas {
				if len(ret.Revert()) > 0 {
					return 0, newRevertError(ret.Revert())
				}
				return 0, ret.Err
			}
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", cap)
		}
	}
	return ethhexutil.Uint64(hi), nil
}

// accessListResult returns an optional accesslist
// It's the result of the `debug_createAccessList` RPC call.
// It contains an error if the transaction itself failed.
type accessListResult struct {
	Accesslist *ethtypes.AccessList `json:"accessList"`
	Error      string               `json:"error,omitempty"`
	GasUsed    ethhexutil.Uint64    `json:"gasUsed"`
}

func (s *BlockChainAPI) CreateAccessList(args types.CallArgs, blockNrOrHash *rpctypes.BlockNumberOrHash) (*accessListResult, error) {
	return nil, ErrNotSupportApiError
}

// FormatBlock creates an ethereum block from a tendermint header and ethereum-formatted
// transactions.
func formatBlock(api api.CoreAPI, genesisConfig *repo.GenesisConfig, block *types.Block, fullTx bool) (map[string]any, error) {
	var err error
	formatTx := func(tx *types.Transaction, index uint64) (any, error) {
		return tx.GetHash().ETHHash(), nil
	}
	if fullTx {
		formatTx = func(tx *types.Transaction, index uint64) (any, error) {
			return NewRPCTransaction(tx, common.BytesToHash(block.BlockHash.Bytes()), block.Height(), index), nil
		}
	}
	txs := block.Transactions
	transactions := make([]any, len(txs))
	for i, tx := range txs {
		if transactions[i], err = formatTx(tx, uint64(i)); err != nil {
			return nil, err
		}
	}

	gasPrice, err := api.Gas().GetCurrentGasPrice(block.Height())
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"number":           (*ethhexutil.Big)(big.NewInt(int64(block.Height()))),
		"hash":             block.BlockHash.ETHHash(),
		"baseFeePerGas":    ethhexutil.Uint64(gasPrice),
		"parentHash":       block.BlockHeader.ParentHash.ETHHash(),
		"nonce":            ethtypes.BlockNonce{}, // PoW specific
		"logsBloom":        block.BlockHeader.Bloom.ETHBloom(),
		"transactionsRoot": block.BlockHeader.TxRoot.ETHHash(),
		"stateRoot":        block.BlockHeader.StateRoot.ETHHash(),
		"miner":            block.BlockHeader.ProposerAccount,
		"extraData":        ethhexutil.Bytes{},
		"size":             ethhexutil.Uint64(block.Size()),
		"gasLimit":         ethhexutil.Uint64(genesisConfig.EpochInfo.FinanceParams.GasLimit), // Static gas limit
		"gasUsed":          ethhexutil.Uint64(block.BlockHeader.GasUsed),
		"timestamp":        ethhexutil.Uint64(block.BlockHeader.Timestamp),
		"transactions":     transactions,
		"receiptsRoot":     block.BlockHeader.ReceiptRoot.ETHHash(),
		// todo delete non-existent fields
		"sha3Uncles": ethtypes.EmptyUncleHash, // No uncles in raft/rbft
		"uncles":     []string{},
		"mixHash":    common.Hash{},
		"difficulty": (*ethhexutil.Big)(big.NewInt(0)),
	}, nil
}

// revertError is an API error that encompassas an EVM revertal with JSON error
// code and a binary data blob.
type revertError struct {
	error
	reason string // revert reason hex encoded
}

// ErrorCode returns the JSON error code for a revertal.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func (e *revertError) ErrorCode() int {
	return 3
}

// ErrorData returns the hex encoded revert reason.
func (e *revertError) ErrorData() any {
	return e.reason
}

func newRevertError(data []byte) *revertError {
	reason, errUnpack := abi.UnpackRevert(data)
	err := errors.New("execution reverted")
	if errUnpack == nil {
		err = fmt.Errorf("execution reverted: %v", reason)
	}
	return &revertError{
		error:  err,
		reason: ethhexutil.Encode(data),
	}
}
