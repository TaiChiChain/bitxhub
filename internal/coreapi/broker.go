package coreapi

import (
	"errors"
	"fmt"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/eth-kit/adaptor"
	vm "github.com/axiomesh/eth-kit/evm"
)

type BrokerAPI CoreAPI

func (b *BrokerAPI) GetTotalPendingTxCount() uint64 {
	if b.axiomLedger.Repo.ReadonlyMode {
		return 0
	}
	return b.axiomLedger.Consensus.GetTotalPendingTxCount()
}

var _ api.BrokerAPI = (*BrokerAPI)(nil)

func (b *BrokerAPI) HandleTransaction(tx *types.Transaction) error {
	if b.axiomLedger.Repo.ReadonlyMode {
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

func (b *BrokerAPI) GetBlock(mode string, value string) (*types.Block, error) {
	switch mode {
	case "HEIGHT":
		height, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("wrong block number: %s", value)
		}
		return b.axiomLedger.ViewLedger.ChainLedger.GetBlock(height)
	case "HASH":
		hash := types.NewHashByStr(value)
		if hash == nil {
			return nil, errors.New("invalid format of block hash for querying block")
		}
		return b.axiomLedger.ViewLedger.ChainLedger.GetBlockByHash(hash)
	default:
		return nil, fmt.Errorf("wrong args about getting block: %s", mode)
	}
}

func (b *BrokerAPI) GetBlocks(start uint64, end uint64) ([]*types.Block, error) {
	meta := b.axiomLedger.ViewLedger.ChainLedger.GetChainMeta()

	var blocks []*types.Block
	if meta.Height < end {
		end = meta.Height
	}
	for i := start; i > 0 && i <= end; i++ {
		b, err := b.GetBlock("HEIGHT", strconv.Itoa(int(i)))
		if err != nil {
			continue
		}
		blocks = append(blocks, b)
	}

	return blocks, nil
}

func (b *BrokerAPI) GetBlockHeaders(start uint64, end uint64) ([]*types.BlockHeader, error) {
	meta := b.axiomLedger.ViewLedger.ChainLedger.GetChainMeta()

	var blockHeaders []*types.BlockHeader
	if meta.Height < end {
		end = meta.Height
	}
	for i := start; i > 0 && i <= end; i++ {
		b, err := b.GetBlock("HEIGHT", strconv.Itoa(int(i)))
		if err != nil {
			continue
		}
		blockHeaders = append(blockHeaders, b.BlockHeader)
	}

	return blockHeaders, nil
}

func (b *BrokerAPI) ConsensusReady() error {
	if b.axiomLedger.Repo.ReadonlyMode {
		return nil
	}

	return b.axiomLedger.Consensus.Ready()
}

func (b *BrokerAPI) GetPendingTxCountByAccount(account string) uint64 {
	if b.axiomLedger.Repo.ReadonlyMode {
		return 0
	}
	return b.axiomLedger.Consensus.GetPendingTxCountByAccount(account)
}

func (b *BrokerAPI) GetPoolTransaction(hash *types.Hash) *types.Transaction {
	if b.axiomLedger.Repo.ReadonlyMode {
		return nil
	}
	return b.axiomLedger.Consensus.GetPendingTxByHash(hash)
}

func (b *BrokerAPI) GetViewStateLedger() ledger.StateLedger {
	return b.axiomLedger.ViewLedger.StateLedger
}

func (b *BrokerAPI) GetEvm(mes *vm.Message, vmConfig *vm.Config) (*vm.EVM, error) {
	if vmConfig == nil {
		vmConfig = new(vm.Config)
	}
	txContext := vm.NewEVMTxContext(mes)
	return b.axiomLedger.BlockExecutor.NewEvmWithViewLedger(txContext, *vmConfig)
}

func (b *BrokerAPI) GetSystemContract(addr *ethcommon.Address) (common.SystemContract, bool) {
	if addr == nil {
		return nil, false
	}
	return system.GetSystemContract(types.NewAddress(addr.Bytes()))
}

func (b *BrokerAPI) StateAtTransaction(block *types.Block, txIndex int, reexec uint64) (*vm.Message, vm.BlockContext, *ledger.StateLedger, error) {
	if block.Height() == 1 {
		return nil, vm.BlockContext{}, nil, errors.New("no transaction in genesis")
	}
	parent, err := b.axiomLedger.ViewLedger.ChainLedger.GetBlockByHash(block.BlockHeader.ParentHash)
	if err != nil || parent == nil {
		return nil, vm.BlockContext{}, nil, fmt.Errorf("parent %#x not found", *block.BlockHeader.ParentHash)
	}

	statedb := b.axiomLedger.ViewLedger.StateLedger.NewView(parent)

	if err != nil {
		return nil, vm.BlockContext{}, nil, err
	}
	if txIndex == 0 && len(block.Transactions) == 0 {
		return nil, vm.BlockContext{}, &statedb, nil
	}

	for idx, tx := range block.Transactions {
		// Assemble the transaction call message and return if the requested offset
		msg := adaptor.TransactionToMessage(tx)
		txContext := vm.NewEVMTxContext(msg)

		context := vm.NewEVMBlockContext(block.Height(), uint64(block.BlockHeader.Timestamp), block.BlockHeader.ProposerAccount, getBlockHashFunc(block))
		if idx == txIndex {
			return msg, context, &statedb, nil
		}
		// Not yet the searched for transaction, execute on top of the current state
		vmenv := vm.NewEVM(context, txContext, statedb, b.axiomLedger.BlockExecutor.GetChainConfig(), vm.Config{})

		statedb.SetTxContext(tx.GetHash(), idx)
		if _, err := vm.ApplyMessage(vmenv, msg, new(vm.GasPool).AddGas(tx.GetGas())); err != nil {
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
		hash := block.BlockHash
		if hash == nil {
			return ethcommon.Hash{}
		}
		return ethcommon.BytesToHash(hash.Bytes())
	}
}
