package coreapi

import (
	"errors"
	"fmt"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/executor"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type BrokerAPI CoreAPI

var _ api.BrokerAPI = (*BrokerAPI)(nil)

func (b *BrokerAPI) HandleTransaction(tx *types.Transaction) error {
	if b.axiomLedger.Repo.StartArgs.ReadonlyMode {
		return errors.New("readonly mode cannot process tx")
	}

	if tx.GetHash() == nil {
		return errors.New("transaction hash is nil")
	}

	b.logger.WithFields(logrus.Fields{
		"hash": tx.GetHash().String(),
	}).Debugf("Receive tx")

	if err := b.axiomLedger.Consensus.Prepare(tx); err != nil {
		return fmt.Errorf("consensus prepare for tx %s failed: %w", tx.GetHash().String(), err)
	}

	return nil
}

func (b *BrokerAPI) GetTransaction(hash *types.Hash) (*types.Transaction, error) {
	return b.axiomLedger.ViewLedger.ChainLedger.GetTransaction(hash)
}

func (b *BrokerAPI) GetTransactionMeta(hash *types.Hash) (*types.TransactionMeta, error) {
	return b.axiomLedger.ViewLedger.ChainLedger.GetTransactionMeta(hash)
}

func (b *BrokerAPI) GetReceipt(hash *types.Hash) (*types.Receipt, error) {
	return b.axiomLedger.ViewLedger.ChainLedger.GetReceipt(hash)
}

func (b *BrokerAPI) GetBlockHeader(mode string, key string) (*types.BlockHeader, error) {
	switch mode {
	case "HEIGHT":
		height, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("wrong block number: %s", key)
		}
		return b.axiomLedger.ViewLedger.ChainLedger.GetBlockHeader(height)
	case "HASH":
		hash := types.NewHashByStr(key)
		if hash == nil {
			return nil, errors.New("invalid format of block hash for querying block")
		}
		number, err := b.axiomLedger.ViewLedger.ChainLedger.GetBlockNumberByHash(hash)
		if err != nil {
			return nil, err
		}
		return b.axiomLedger.ViewLedger.ChainLedger.GetBlockHeader(number)
	default:
		return nil, fmt.Errorf("wrong args about getting block: %s", mode)
	}
}

func (b *BrokerAPI) GetBlockExtra(height uint64) (*types.BlockExtra, error) {
	return b.axiomLedger.ViewLedger.ChainLedger.GetBlockExtra(height)
}

func (b *BrokerAPI) GetBlockTxHashList(height uint64) ([]*types.Hash, error) {
	return b.axiomLedger.ViewLedger.ChainLedger.GetBlockTxHashList(height)
}

func (b *BrokerAPI) GetBlockTxList(height uint64) ([]*types.Transaction, error) {
	return b.axiomLedger.ViewLedger.ChainLedger.GetBlockTxList(height)
}

func (b *BrokerAPI) ConsensusReady() error {
	if b.axiomLedger.Repo.StartArgs.ReadonlyMode {
		return nil
	}

	return b.axiomLedger.Consensus.Ready()
}

func (b *BrokerAPI) GetViewStateLedger() ledger.StateLedger {
	return b.axiomLedger.ViewLedger.StateLedger
}

func (b *BrokerAPI) GetEvm(mes *core.Message, vmConfig *vm.Config) (*vm.EVM, error) {
	if vmConfig == nil {
		vmConfig = new(vm.Config)
	}
	txContext := core.NewEVMTxContext(mes)
	return b.axiomLedger.BlockExecutor.NewEvmWithViewLedger(txContext, *vmConfig)
}

func (b *BrokerAPI) GetNativeVm() common.VirtualMachine {
	return b.axiomLedger.BlockExecutor.NewViewNativeVM()
}

func (b *BrokerAPI) StateAtTransaction(block *types.Block, txIndex int, reexec uint64) (*core.Message, vm.BlockContext, *ledger.StateLedger, error) {
	if block.Height() == b.axiomLedger.Repo.GenesisConfig.EpochInfo.StartBlock {
		return nil, vm.BlockContext{}, nil, errors.New("no transaction in genesis")
	}

	parentNumber, err := b.axiomLedger.ViewLedger.ChainLedger.GetBlockNumberByHash(block.Header.ParentHash)
	if err != nil {
		return nil, vm.BlockContext{}, nil, err
	}
	parentHeader, err := b.axiomLedger.ViewLedger.ChainLedger.GetBlockHeader(parentNumber)
	if err != nil || parentHeader == nil {
		return nil, vm.BlockContext{}, nil, fmt.Errorf("parent %#x not found", block.Header.ParentHash)
	}

	statedb := b.axiomLedger.ViewLedger.StateLedger.NewViewWithoutCache(parentHeader, false)
	if txIndex == 0 && len(block.Transactions) == 0 {
		return nil, vm.BlockContext{}, &statedb, nil
	}

	for idx, tx := range block.Transactions {
		// Assemble the transaction call message and return if the requested offset
		msg := executor.TransactionToMessage(tx)
		txContext := core.NewEVMTxContext(msg)

		context := executor.NewEVMBlockContextAdaptor(block.Height(), uint64(block.Header.Timestamp), block.Header.ProposerAccount, getBlockHashFunc(block))
		if idx == txIndex {
			return msg, context, &statedb, nil
		}
		// Not yet the searched for transaction, execute on top of the current state
		vmenv := vm.NewEVM(context, txContext, &ledger.EvmStateDBAdaptor{StateLedger: statedb}, b.axiomLedger.BlockExecutor.GetChainConfig(), vm.Config{})

		statedb.SetTxContext(tx.GetHash(), idx)
		if _, err := core.ApplyMessage(vmenv, msg, new(core.GasPool).AddGas(tx.GetGas())); err != nil {
			return nil, vm.BlockContext{}, nil, fmt.Errorf("transaction %#x failed: %v", tx.GetHash(), err)
		}
		// Ensure any modifications are committed to the state
		// Only delete empty objects if EIP158/161 (a.k.a Spurious Dragon) is in effect
		statedb.Finalise()
	}
	return nil, vm.BlockContext{}, nil, fmt.Errorf("transaction index %d out of range for block %#x", txIndex, block.Hash())
}

func (b *BrokerAPI) ChainConfig() *params.ChainConfig {
	return b.axiomLedger.BlockExecutor.GetChainConfig()
}

func getBlockHashFunc(block *types.Block) vm.GetHashFunc {
	return func(n uint64) ethcommon.Hash {
		hash := block.Hash()
		if hash == nil {
			return ethcommon.Hash{}
		}
		return ethcommon.BytesToHash(hash.Bytes())
	}
}
