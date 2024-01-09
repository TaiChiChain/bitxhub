package executor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/axiomesh/axiom-ledger/internal/executor/system"

	"github.com/cbergoon/merkletree"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	consensuscommon "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	sys_common "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/axiomesh/eth-kit/adaptor"
	ethvm "github.com/axiomesh/eth-kit/evm"
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
		exec.logger.WithFields(logrus.Fields{
			"begin height": newBlock.Height() - 1,
			"end height":   exec.currentHeight,
			"err":          err.Error(),
		}).Errorf("rollback block error")
		return err
	}

	// query last checked block for generating right parent blockHash
	lastCheckedBlock, err := exec.ledger.ChainLedger.GetBlock(newBlock.Height() - 1)
	if err != nil {
		exec.logger.WithFields(logrus.Fields{
			"height": lastCheckedBlock.Height(),
			"err":    err.Error(),
		}).Errorf("get last checked block from ledger error")
		return err
	}
	// rollback currentHeight and currentBlockHash
	exec.currentHeight = newBlock.Height() - 1
	exec.currentBlockHash = lastCheckedBlock.BlockHash

	exec.logger.WithFields(logrus.Fields{
		"height": lastCheckedBlock.Height(),
		"hash":   lastCheckedBlock.BlockHash.String(),
	}).Infof("rollback block success")

	return nil
}

func (exec *BlockExecutor) processExecuteEvent(commitEvent *consensuscommon.CommitEvent) {
	var txHashList []*types.Hash
	current := time.Now()
	block := commitEvent.Block

	// check executor handle the right block
	if block.BlockHeader.Number != exec.currentHeight+1 {
		exec.logger.WithFields(logrus.Fields{"block height": block.BlockHeader.Number,
			"matchedHeight": exec.currentHeight + 1}).Warning("current block height is not matched")
		if block.BlockHeader.Number <= exec.currentHeight {
			if exec.rep.Config.Executor.DisableRollback {
				panic(fmt.Sprintf("not supported rollback to %d", block.BlockHeader.Number))
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
	exec.evm = newEvm(exec.rep.Config.Executor.EVM, block.Height(), uint64(block.BlockHeader.Timestamp), exec.evmChainCfg, exec.ledger.StateLedger, exec.ledger.ChainLedger, block.BlockHeader.ProposerAccount)
	// get last block's stateRoot to init the latest world state trie
	parentBlock, err := exec.ledger.ChainLedger.GetBlock(block.Height() - 1)
	if err != nil {
		exec.logger.WithFields(logrus.Fields{
			"height": block.Height() - 1,
			"err":    err.Error(),
		}).Errorf("get last block from ledger error")
		panic(err)
	}
	exec.ledger.StateLedger.PrepareBlock(parentBlock.BlockHeader.StateRoot, block.BlockHash, block.Height())
	receipts := exec.applyTransactions(block.Transactions, block.Height())

	// check need turn into NewEpoch
	epochInfo := exec.rep.EpochInfo
	epochChange := false
	if block.BlockHeader.Number == (epochInfo.StartBlock + epochInfo.EpochPeriod - 1) {
		var seed []byte
		seed = append(seed, []byte(exec.currentBlockHash.String())...)
		seed = append(seed, []byte(block.BlockHeader.ProposerAccount)...)
		seed = binary.BigEndian.AppendUint64(seed, block.BlockHeader.Number)
		seed = binary.BigEndian.AppendUint64(seed, block.BlockHeader.Epoch)
		seed = binary.BigEndian.AppendUint64(seed, uint64(block.BlockHeader.Timestamp))
		for _, tx := range block.Transactions {
			seed = append(seed, []byte(tx.GetHash().String())...)
		}

		newEpoch, err := base.TurnIntoNewEpoch(seed, exec.ledger.StateLedger)
		if err != nil {
			panic(err)
		}
		exec.rep.EpochInfo = newEpoch
		epochChange = true
		exec.logger.WithFields(logrus.Fields{
			"height":                commitEvent.Block.BlockHeader.Number,
			"new_epoch":             newEpoch.Epoch,
			"new_epoch_start_block": newEpoch.StartBlock,
		}).Info("Turn into new epoch")
		exec.ledger.StateLedger.Finalise()
	}

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

	block.BlockHeader.TxRoot = txRoot
	block.BlockHeader.ReceiptRoot = receiptRoot
	block.BlockHeader.ParentHash = exec.currentBlockHash

	parentChainMeta := exec.ledger.ChainLedger.GetChainMeta()

	var nextGasPrice uint64
	if epochChange && exec.rep.EpochInfo.FinanceParams.StartGasPriceAvailable {
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

	block.BlockHeader.GasPrice = int64(nextGasPrice)

	stateRoot, err := exec.ledger.StateLedger.Commit()
	if err != nil {
		panic(fmt.Errorf("commit stateLedger failed: %w", err))
	}

	block.BlockHeader.StateRoot = stateRoot
	block.BlockHeader.GasUsed = exec.cumulativeGasUsed
	block.BlockHash = block.Hash()

	exec.logger.WithFields(logrus.Fields{
		"hash":             block.BlockHash.String(),
		"height":           block.BlockHeader.Number,
		"epoch":            block.BlockHeader.Epoch,
		"coinbase":         block.BlockHeader.ProposerAccount,
		"proposer_node_id": block.BlockHeader.ProposerNodeID,
		"gas_price":        block.BlockHeader.GasPrice,
		"gas_used":         block.BlockHeader.GasUsed,
		"parent_hash":      block.BlockHeader.ParentHash.String(),
		"tx_root":          block.BlockHeader.TxRoot.String(),
		"receipt_root":     block.BlockHeader.ReceiptRoot.String(),
		"state_root":       block.BlockHeader.StateRoot.String(),
	}).Info("Block meta")

	calcBlockSize.Observe(float64(block.Size()))
	executeBlockDuration.Observe(float64(time.Since(current)) / float64(time.Second))

	exec.getLogsForReceipt(receipts, block.BlockHash)
	block.BlockHeader.Bloom = ledger.CreateBloom(receipts)

	data := &ledger.BlockData{
		Block:      block,
		Receipts:   receipts,
		TxHashList: txHashList,
	}

	exec.logger.WithFields(logrus.Fields{
		"height": commitEvent.Block.BlockHeader.Number,
		"count":  len(commitEvent.Block.Transactions),
		"elapse": time.Since(current),
	}).Info("Executed block")

	now := time.Now()
	exec.ledger.PersistBlockData(data)

	// metrics for cal tx tps
	txCounter.Add(float64(len(data.Block.Transactions)))
	if block.BlockHeader.ProposerAccount == exec.rep.AccountAddress {
		proposedBlockCounter.Inc()
	}

	exec.logger.WithFields(logrus.Fields{
		"gasPrice": data.Block.BlockHeader.GasPrice,
		"height":   data.Block.BlockHeader.Number,
		"hash":     data.Block.BlockHash.String(),
		"count":    len(data.Block.Transactions),
		"elapse":   time.Since(now),
	}).Info("Persisted block")

	exec.currentHeight = block.BlockHeader.Number
	exec.currentBlockHash = block.BlockHash

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

	var result *ethvm.ExecutionResult
	var err error

	msg := adaptor.TransactionToMessage(tx)
	curGasPrice := exec.getCurrentGasPrice(height)
	if msg.GasPrice.Cmp(curGasPrice) < 0 {
		exec.logger.Warnf("msg gas price %v not equals to current gas price %v, will adjust msg.GasPrice automatically", msg.GasPrice, curGasPrice)
		msg.GasPrice = curGasPrice
	}

	statedb := exec.ledger.StateLedger
	// TODO: Move to system contract
	snapshot := statedb.Snapshot()

	if exec.nvm.IsSystemContract(tx.GetTo()) {
		result = system.RunAxiomNativeVM(exec.nvm, height, statedb, msg.Data, msg.From, msg.To)
		if result != nil && result.UsedGas != 0 {
			fee := new(big.Int).SetUint64(result.UsedGas)
			fee.Mul(fee, msg.GasPrice)
			ethvm.Transfer(statedb, msg.From, exec.evm.Context.Coinbase, fee)
		}
	} else {
		// execute evm
		gp := new(ethvm.GasPool).AddGas(exec.gasLimit)
		txContext := ethvm.NewEVMTxContext(msg)
		exec.evm.Reset(txContext, exec.ledger.StateLedger)
		result, err = ethvm.ApplyMessage(exec.evm, msg, gp)
	}

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
		if strings.HasPrefix(result.Err.Error(), ethvm.ErrExecutionReverted.Error()) {
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
	receipt.EvmLogs = exec.ledger.StateLedger.GetLogs(*receipt.TxHash, height, &types.Hash{})
	receipt.Bloom = ledger.CreateBloom(ledger.EvmReceipts{receipt})
	exec.cumulativeGasUsed += receipt.GasUsed
	receipt.CumulativeGasUsed = exec.cumulativeGasUsed
	receipt.EffectiveGasPrice = 0

	return receipt
}

func (exec *BlockExecutor) clear() {
	exec.ledger.StateLedger.Clear()
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

func getBlockHashFunc(chainLedger ledger.ChainLedger) ethvm.GetHashFunc {
	return func(n uint64) common.Hash {
		hash := chainLedger.GetBlockHash(n)
		if hash == nil {
			return common.Hash{}
		}
		return common.BytesToHash(hash.Bytes())
	}
}

func newEvm(evmCfg repo.EVM, number uint64, timestamp uint64, chainCfg *params.ChainConfig, db ledger.StateLedger, chainLedger ledger.ChainLedger, coinbase string) *ethvm.EVM {
	if coinbase == "" {
		coinbase = sys_common.ZeroAddress
	}

	blkCtx := ethvm.NewEVMBlockContext(number, timestamp, coinbase, getBlockHashFunc(chainLedger))

	return ethvm.NewEVM(blkCtx, ethvm.TxContext{}, db, chainCfg, ethvm.Config{
		DisableMaxCodeSizeLimit: evmCfg.DisableMaxCodeSizeLimit,
	})
}

func (exec *BlockExecutor) NewEvmWithViewLedger(txCtx ethvm.TxContext, vmConfig ethvm.Config) (*ethvm.EVM, error) {
	var blkCtx ethvm.BlockContext
	meta := exec.ledger.ChainLedger.GetChainMeta()
	block, err := exec.ledger.ChainLedger.GetBlock(meta.Height)
	if err != nil {
		exec.logger.Errorf("fail to get block at %d: %v", meta.Height, err.Error())
		return nil, err
	}

	lg := exec.ledger.NewView()
	blkCtx = ethvm.NewEVMBlockContext(meta.Height, uint64(block.BlockHeader.Timestamp), exec.rep.AccountAddress, getBlockHashFunc(exec.ledger.ChainLedger))
	return ethvm.NewEVM(blkCtx, txCtx, lg.StateLedger, exec.evmChainCfg, vmConfig), nil
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

func (exec *BlockExecutor) getLogsForReceipt(receipts []*types.Receipt, hash *types.Hash) {
	for _, receipt := range receipts {
		for _, log := range receipt.EvmLogs {
			log.BlockHash = hash
		}
	}
}
