package executor

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cbergoon/merkletree"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	consensuscommon "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	sys_common "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/events"
)

type InvalidReason string

type BlockWrapper struct {
	block     *types.Block
	invalidTx map[int]InvalidReason
}

func (exec *BlockExecutor) applyTransactions(txs []*types.Transaction, height uint64) []*types.Receipt {
	receipts := make([]*types.Receipt, 0, len(txs))

	for i, tx := range txs {
		receipts = append(receipts, exec.applyTransaction(i, tx, height))
	}

	exec.logger.Debugf("executor executed %d txs", len(txs))

	return receipts
}

func (exec *BlockExecutor) rollbackBlocks(newBlock *types.Block) error {
	// rollback from stateLedger、chainLedger and blockFile
	err := exec.ledger.Rollback(newBlock.Height() - 1)
	if err != nil {
		return errors.Wrapf(err, "rollback block error, begin height: %d, end height: %d", newBlock.Height()-1, exec.currentHeight)
	}

	// query last checked block for generating right parent blockHash
	lastCheckedBlockHeader, err := exec.ledger.ChainLedger.GetBlockHeader(newBlock.Height() - 1)
	if err != nil {
		return errors.Wrapf(err, "get last checked block from ledger error, height: %d", lastCheckedBlockHeader.Number)
	}
	// rollback currentHeight and currentBlockHash
	exec.currentHeight = newBlock.Height() - 1
	exec.currentBlockHash = lastCheckedBlockHeader.Hash()

	exec.logger.WithFields(logrus.Fields{
		"height": lastCheckedBlockHeader.Number,
		"hash":   lastCheckedBlockHeader.Hash().String(),
	}).Infof("rollback block success")

	return nil
}

func (exec *BlockExecutor) processExecuteEvent(commitEvent *consensuscommon.CommitEvent) {
	var txHashList []*types.Hash
	current := time.Now()
	block := commitEvent.Block

	// check executor handle the right block
	if block.Header.Number != exec.currentHeight+1 {
		exec.logger.WithFields(logrus.Fields{"block height": block.Header.Number,
			"matchedHeight": exec.currentHeight + 1}).Warning("current block height is not matched")
		if block.Header.Number <= exec.currentHeight {
			if exec.rep.Config.Executor.DisableRollback {
				panic(fmt.Sprintf("not supported rollback to %d", block.Header.Number))
			}
			err := exec.rollbackBlocks(block)
			if err != nil {
				exec.logger.WithError(err).Error("rollback blocks failed")
				panic(err)
			}
		} else {
			return
		}
	}

	for _, tx := range block.Transactions {
		txHashList = append(txHashList, tx.GetHash())
	}

	exec.cumulativeGasUsed = 0
	exec.evm = newEvm(block.Height(), uint64(block.Header.Timestamp), exec.evmChainCfg, exec.ledger.StateLedger, exec.ledger.ChainLedger, block.Header.ProposerAccount)
	// get last block's stateRoot to init the latest world state trie
	parentBlockHeader, err := exec.ledger.ChainLedger.GetBlockHeader(block.Height() - 1)
	if err != nil {
		exec.logger.WithFields(logrus.Fields{
			"height": block.Height() - 1,
			"err":    err.Error(),
		}).Errorf("get last block from ledger error")
		panic(err)
	}
	exec.ledger.StateLedger.PrepareBlock(parentBlockHeader.StateRoot, block.Height())
	receipts := exec.applyTransactions(block.Transactions, block.Height())

	for _, hook := range exec.afterBlockHooks {
		hook(block)
	}

	exec.ledger.StateLedger.Finalise()

	applyTxsDuration.Observe(float64(time.Since(current)) / float64(time.Second))
	exec.logger.WithFields(logrus.Fields{
		"time":  time.Since(current),
		"count": len(block.Transactions),
	}).Debug("Apply transactions elapsed")

	calcMerkleStart := time.Now()
	txRoot, err := exec.buildTxMerkleTree(block.Transactions)
	if err != nil {
		panic(err)
	}

	receiptRoot, err := exec.calcReceiptMerkleRoot(receipts)
	if err != nil {
		panic(err)
	}

	calcMerkleDuration.Observe(float64(time.Since(calcMerkleStart)) / float64(time.Second))

	block.Header.TxRoot = txRoot
	block.Header.ReceiptRoot = receiptRoot
	block.Header.ParentHash = exec.currentBlockHash

	parentChainMeta := exec.ledger.ChainLedger.GetChainMeta()

	var nextGasPrice uint64
	if exec.epochExchange && exec.rep.EpochInfo.FinanceParams.StartGasPriceAvailable {
		// epoch changed gas price
		nextGasPrice = exec.rep.EpochInfo.FinanceParams.StartGasPrice
		exec.logger.WithFields(logrus.Fields{
			"price": nextGasPrice,
		}).Info("Next epoch start block gas price changed")
	} else {
		nextGasPrice, err = exec.gas.CalNextGasPrice(parentChainMeta.GasPrice.Uint64(), len(block.Transactions))
		if err != nil {
			panic(fmt.Errorf("calculate next block gas price failed: %w", err))
		}
	}

	block.Header.GasPrice = int64(nextGasPrice)

	stateRoot, err := exec.ledger.StateLedger.Commit()
	if err != nil {
		panic(fmt.Errorf("commit stateLedger failed: %w", err))
	}

	block.Header.StateRoot = stateRoot
	block.Header.GasUsed = exec.cumulativeGasUsed

	// update block hash cache
	block.Header.CalculateHash()

	exec.logger.WithFields(logrus.Fields{
		"hash":             block.Hash().String(),
		"height":           block.Header.Number,
		"epoch":            block.Header.Epoch,
		"coinbase":         block.Header.ProposerAccount,
		"proposer_node_id": block.Header.ProposerNodeID,
		"gas_price":        block.Header.GasPrice,
		"gas_used":         block.Header.GasUsed,
		"parent_hash":      block.Header.ParentHash.String(),
		"tx_root":          block.Header.TxRoot.String(),
		"receipt_root":     block.Header.ReceiptRoot.String(),
		"state_root":       block.Header.StateRoot.String(),
	}).Info("Block meta")

	calcBlockSize.Observe(float64(block.Size()))
	executeBlockDuration.Observe(float64(time.Since(current)) / float64(time.Second))

	exec.updateLogsBlockHash(receipts, block.Hash())
	block.Header.Bloom = ledger.CreateBloom(receipts)

	data := &ledger.BlockData{
		Block:      block,
		Receipts:   receipts,
		TxHashList: txHashList,
	}

	exec.logger.WithFields(logrus.Fields{
		"height": commitEvent.Block.Header.Number,
		"count":  len(commitEvent.Block.Transactions),
		"elapse": time.Since(current),
	}).Info("Executed block")

	now := time.Now()
	exec.ledger.PersistBlockData(data)

	// metrics for cal tx tps
	txCounter.Add(float64(len(data.Block.Transactions)))
	if block.Header.ProposerAccount == exec.rep.P2PAddress {
		proposedBlockCounter.Inc()
	}

	exec.logger.WithFields(logrus.Fields{
		"gasPrice": data.Block.Header.GasPrice,
		"height":   data.Block.Header.Number,
		"hash":     data.Block.Hash().String(),
		"count":    len(data.Block.Transactions),
		"elapse":   time.Since(now),
	}).Info("Persisted block")

	exec.currentHeight = block.Header.Number
	exec.currentBlockHash = block.Hash()

	txPointerList := make([]*events.TxPointer, len(data.Block.Transactions))
	lo.ForEach(data.Block.Transactions, func(item *types.Transaction, index int) {
		txPointerList[index] = &events.TxPointer{
			Hash:    item.GetHash(),
			Account: item.RbftGetFrom(),
			Nonce:   item.RbftGetNonce(),
		}
	})

	exec.postBlockEvent(data.Block, txPointerList, commitEvent.StateUpdatedCheckpoint)
	exec.postLogsEvent(data.Receipts)
	exec.clear()
}

func (exec *BlockExecutor) buildTxMerkleTree(txs []*types.Transaction) (*types.Hash, error) {
	hash, err := calcMerkleRoot(lo.Map(txs, func(item *types.Transaction, index int) merkletree.Content {
		return item.GetHash()
	}))
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (exec *BlockExecutor) postBlockEvent(block *types.Block, txPointerList []*events.TxPointer, ckp *consensus.Checkpoint) {
	exec.blockFeed.Send(events.ExecutedEvent{
		Block:                  block,
		TxPointerList:          txPointerList,
		StateUpdatedCheckpoint: ckp,
	})
	exec.blockFeedForRemote.Send(events.ExecutedEvent{
		Block:         block,
		TxPointerList: txPointerList,
	})
}

func (exec *BlockExecutor) postLogsEvent(receipts []*types.Receipt) {
	logs := make([]*types.EvmLog, 0)
	for _, receipt := range receipts {
		logs = append(logs, receipt.EvmLogs...)
	}

	exec.logsFeed.Send(logs)
}

func (exec *BlockExecutor) applyTransaction(i int, tx *types.Transaction, height uint64) *types.Receipt {
	defer func() {
		exec.ledger.StateLedger.SetNonce(tx.GetFrom(), tx.GetNonce()+1)
		exec.ledger.StateLedger.Finalise()
	}()

	exec.ledger.StateLedger.SetTxContext(tx.GetHash(), i)

	receipt := &types.Receipt{
		TxHash: tx.GetHash(),
	}

	var result *core.ExecutionResult
	var err error

	msg := TransactionToMessage(tx)
	curGasPrice := exec.getCurrentGasPrice(height)
	msg.GasPrice = curGasPrice
	msg.GasFeeCap = curGasPrice
	msg.GasTipCap = curGasPrice

	statedb := exec.ledger.StateLedger
	evmStateDB := &ledger.EvmStateDBAdaptor{StateLedger: statedb}
	// TODO: Move to system contract
	snapshot := statedb.Snapshot()

	// execute evm
	gp := new(core.GasPool).AddGas(exec.gasLimit)
	txContext := core.NewEVMTxContext(msg)
	exec.evm.Reset(txContext, evmStateDB)
	result, err = core.ApplyMessage(exec.evm, msg, gp)

	if err != nil {
		exec.logger.Errorf("apply tx failed: %s", err.Error())
		statedb.RevertToSnapshot(snapshot)
		receipt.Status = types.ReceiptFAILED
		receipt.Ret = []byte(err.Error())
		return receipt
	}
	if result.Failed() {
		if len(result.Revert()) > 0 {
			reason, errUnpack := abi.UnpackRevert(result.Revert())
			if errUnpack == nil {
				exec.logger.Warnf("execute tx failed: %s: %s", result.Err.Error(), reason)
			} else {
				exec.logger.Warnf("execute tx failed: %s", result.Err.Error())
			}
		} else {
			exec.logger.Warnf("execute tx failed: %s", result.Err.Error())
		}

		receipt.Status = types.ReceiptFAILED
		receipt.Ret = []byte(result.Err.Error())
		if strings.HasPrefix(result.Err.Error(), vm.ErrExecutionReverted.Error()) {
			receipt.Ret = append(receipt.Ret, common.CopyBytes(result.ReturnData)...)
		}
	} else {
		receipt.Status = types.ReceiptSUCCESS
		receipt.Ret = result.Return()
	}

	receipt.TxHash = tx.GetHash()
	receipt.GasUsed = result.UsedGas
	if msg.To == nil || bytes.Equal(msg.To.Bytes(), common.Address{}.Bytes()) {
		receipt.ContractAddress = types.NewAddress(crypto.CreateAddress(exec.evm.TxContext.Origin, tx.GetNonce()).Bytes())
	}
	receipt.EvmLogs = exec.ledger.StateLedger.GetLogs(*receipt.TxHash, height)
	receipt.Bloom = ledger.CreateBloom(ledger.EvmReceipts{receipt})
	exec.cumulativeGasUsed += receipt.GasUsed
	receipt.CumulativeGasUsed = exec.cumulativeGasUsed
	receipt.EffectiveGasPrice = 0

	return receipt
}

func (exec *BlockExecutor) clear() {
	exec.ledger.StateLedger.Clear()
	exec.epochExchange = false
}

func (exec *BlockExecutor) calcReceiptMerkleRoot(receipts []*types.Receipt) (*types.Hash, error) {
	current := time.Now()

	receiptHashes := make([]merkletree.Content, 0, len(receipts))
	for _, receipt := range receipts {
		receiptHashes = append(receiptHashes, receipt.Hash())
	}
	receiptRoot, err := calcMerkleRoot(receiptHashes)
	if err != nil {
		return nil, err
	}

	exec.logger.WithField("time", time.Since(current)).Debug("Calculate receipt merkle roots")

	return receiptRoot, nil
}

func calcMerkleRoot(contents []merkletree.Content) (*types.Hash, error) {
	if len(contents) == 0 {
		// compatible with Ethereum
		return types.NewHashByStr("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), nil
	}

	tree, err := merkletree.NewTree(contents)
	if err != nil {
		return nil, err
	}

	return types.NewHash(tree.MerkleRoot()), nil
}

func getBlockHashFunc(chainLedger ledger.ChainLedger) vm.GetHashFunc {
	return func(n uint64) common.Hash {
		blockHeader, err := chainLedger.GetBlockHeader(n)
		if err != nil {
			return common.Hash{}
		}
		return common.BytesToHash(blockHeader.Hash().Bytes())
	}
}

func newEvm(number uint64, timestamp uint64, chainCfg *params.ChainConfig, db ledger.StateLedger, chainLedger ledger.ChainLedger, coinbase string) *vm.EVM {
	if coinbase == "" {
		coinbase = sys_common.ZeroAddress
	}

	blkCtx := NewEVMBlockContextAdaptor(number, timestamp, coinbase, getBlockHashFunc(chainLedger))

	return vm.NewEVM(blkCtx, vm.TxContext{}, &ledger.EvmStateDBAdaptor{
		StateLedger: db,
	}, chainCfg, vm.Config{})
}

func (exec *BlockExecutor) NewEvmWithViewLedger(txCtx vm.TxContext, vmConfig vm.Config) (*vm.EVM, error) {
	var blkCtx vm.BlockContext
	meta := exec.ledger.ChainLedger.GetChainMeta()
	blockHeader, err := exec.ledger.ChainLedger.GetBlockHeader(meta.Height)
	if err != nil {
		return nil, err
	}

	evmLg := &ledger.EvmStateDBAdaptor{
		StateLedger: exec.ledger.NewView().StateLedger,
	}
	blkCtx = NewEVMBlockContextAdaptor(meta.Height, uint64(blockHeader.Timestamp), blockHeader.ProposerAccount, getBlockHashFunc(exec.ledger.ChainLedger))
	return vm.NewEVM(blkCtx, txCtx, evmLg, exec.evmChainCfg, vmConfig), nil
}

func (exec *BlockExecutor) GetChainConfig() *params.ChainConfig {
	return exec.evmChainCfg
}

func (exec *BlockExecutor) NewViewNativeVM() sys_common.VirtualMachine {
	return exec.nvm.View()
}

// getCurrentGasPrice returns the current block's gas price, which is
// stored in the last block's blockheader
func (exec *BlockExecutor) getCurrentGasPrice(height uint64) *big.Int {
	// epoch changed gas price
	if height == exec.rep.EpochInfo.StartBlock && exec.rep.EpochInfo.FinanceParams.StartGasPriceAvailable {
		return big.NewInt(int64(exec.rep.EpochInfo.FinanceParams.StartGasPrice))
	}
	return exec.ledger.ChainLedger.GetChainMeta().GasPrice
}

func (exec *BlockExecutor) updateLogsBlockHash(receipts []*types.Receipt, hash *types.Hash) {
	for _, receipt := range receipts {
		for _, log := range receipt.EvmLogs {
			log.BlockHash = hash
		}
	}
}
