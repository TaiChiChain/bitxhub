package eth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types3 "github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// TransactionAPI provide apis to get and create transaction
type TransactionAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

func NewTransactionAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *TransactionAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &TransactionAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by its height.
func (api *TransactionAPI) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
	}(time.Now())

	api.logger.Debugf("eth_getBlockTransactionCountByNumber, block number: %d", blockNum)
	if blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber {
		meta, _ := api.api.Chain().Meta()
		blockNum = rpctypes.BlockNumber(meta.Height)
	}

	block, err := api.api.Broker().GetBlock("HEIGHT", fmt.Sprintf("%d", blockNum))
	if err != nil {
		queryFailedCounter.Inc()
		api.logger.Debugf("eth api GetBlockTransactionCountByNumber err:%s", err)
		return nil
	}

	count := uint(len(block.Transactions))

	return (*hexutil.Uint)(&count)
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (api *TransactionAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
	}(time.Now())

	api.logger.Debugf("eth_getBlockTransactionCountByHash, hash: %s", hash.String())

	block, err := api.api.Broker().GetBlock("HASH", hash.String())
	if err != nil {
		api.logger.Debugf("eth api GetBlockTransactionCountByHash err:%s", err)
		queryFailedCounter.Inc()
		return nil
	}

	count := uint(len(block.Transactions))

	return (*hexutil.Uint)(&count)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (api *TransactionAPI) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (ret *rpctypes.RPCTransaction, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getTransactionByBlockNumberAndIndex, number: %d, index: %d", blockNum, idx)

	height := uint64(0)

	switch blockNum {
	// Latest and Pending type return current block height
	case rpctypes.LatestBlockNumber, rpctypes.PendingBlockNumber:
		meta, err := api.api.Chain().Meta()
		if err != nil {
			return nil, err
		}
		height = meta.Height
	default:
		height = uint64(blockNum.Int64())
	}

	return getTxByBlockInfoAndIndex(api.api, "HEIGHT", fmt.Sprintf("%d", height), idx)
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (api *TransactionAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (ret *rpctypes.RPCTransaction, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getTransactionByHashAndIndex, hash: %s, index: %d", hash.String(), idx)

	return getTxByBlockInfoAndIndex(api.api, "HASH", hash.String(), idx)
}

// GetTransactionCount returns the number of transactions at the given address, blockNum is ignored.
func (api *TransactionAPI) GetTransactionCount(address common.Address, blockNrOrHash *rpctypes.BlockNumberOrHash) (ret *hexutil.Uint64, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getTransactionCount, address: %s, blockNrOrHash: %v", address, blockNrOrHash)
	if blockNrOrHash != nil {
		if blockNumber, ok := blockNrOrHash.Number(); ok && blockNumber == rpctypes.PendingBlockNumber {
			nonce := api.api.TxPool().GetPendingTxCountByAccount(address.String())
			return (*hexutil.Uint64)(&nonce), nil
		}
	}
	api.logger.Debugf("eth_getTransactionCount from ledger")
	stateLedger, err := getStateLedgerAt(api.api, blockNrOrHash)
	if err != nil {
		return nil, err
	}

	nonce := stateLedger.GetNonce(types.NewAddress(address.Bytes()))

	return (*hexutil.Uint64)(&nonce), nil
}

// GetTransactionByHash returns the transaction identified by hash.
func (api *TransactionAPI) GetTransactionByHash(hash common.Hash) (ret *rpctypes.RPCTransaction, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getTransactionByHash, hash: %s", hash.String())

	typesHash := types.NewHash(hash.Bytes())
	tx, err := api.api.Broker().GetTransaction(typesHash)
	if err != nil && err != storage.ErrorNotFound {
		return nil, err
	}
	if tx != nil {
		meta, err := api.api.Broker().GetTransactionMeta(typesHash)
		if err != nil {
			return nil, fmt.Errorf("get tx meta from ledger: %w", err)
		}
		return NewRPCTransaction(tx, common.BytesToHash(meta.BlockHash.Bytes()), meta.BlockHeight, meta.Index), nil
	}

	// retrieve tx from the pool
	if poolTx := api.api.TxPool().GetTransaction(typesHash); poolTx != nil {
		return NewRPCTransaction(poolTx, common.Hash{}, 0, 0), nil
	}

	return nil, nil
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (api *TransactionAPI) GetTransactionReceipt(hash common.Hash) (ret map[string]any, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debugf("eth_getTransactionReceipt, hash: %s", hash.String())

	txHash := types.NewHash(hash.Bytes())
	// tx, meta, err := getEthTransactionByHash(api.rep, api.api, api.logger, txHash)
	tx, err := api.api.Broker().GetTransaction(txHash)
	if err != nil {
		return nil, nil
	}

	meta, err := api.api.Broker().GetTransactionMeta(txHash)
	if err != nil {
		return nil, fmt.Errorf("get tx meta from ledger: %w", err)
	}
	if err != nil {
		api.logger.Debugf("no tx found for hash %s", txHash.String())
		return nil, err
	}

	receipt, err := api.api.Broker().GetReceipt(txHash)
	if err != nil {
		api.logger.Debugf("no receipt found for tx %s", txHash.String())
		return nil, err
	}

	fields := map[string]any{
		"type":              hexutil.Uint(tx.GetType()),
		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"transactionHash":   hash,
		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		"blockHash":         common.BytesToHash(meta.BlockHash.Bytes()),
		"blockNumber":       hexutil.Uint64(meta.BlockHeight),
		"transactionIndex":  hexutil.Uint64(meta.Index),
		"from":              common.BytesToAddress(tx.GetFrom().Bytes()),
	}
	if receipt.Bloom == nil {
		emptyBloom := types.Bloom{}
		fields["logsBloom"] = emptyBloom.ETHBloom()
	} else {
		fields["logsBloom"] = receipt.Bloom.ETHBloom()
	}
	ethLogs := make([]*types3.Log, 0)
	for _, log := range receipt.EvmLogs {
		ethLog := &types3.Log{
			Address: log.Address.ETHAddress(),
			Topics: lo.Map(log.Topics, func(item *types.Hash, index int) common.Hash {
				return item.ETHHash()
			}),
			Data:        log.Data,
			BlockNumber: log.BlockNumber,
			TxHash:      log.TransactionHash.ETHHash(),
			TxIndex:     uint(log.TransactionIndex),
			BlockHash:   log.BlockHash.ETHHash(),
			Index:       uint(log.LogIndex),
			Removed:     log.Removed,
		}
		ethLogs = append(ethLogs, ethLog)
	}
	fields["logs"] = ethLogs

	if receipt.Status == types.ReceiptSUCCESS {
		fields["status"] = hexutil.Uint(1)
	} else {
		fields["status"] = hexutil.Uint(0)
	}

	if receipt.ContractAddress != nil {
		fields["contractAddress"] = common.BytesToAddress(receipt.ContractAddress.Bytes())
	} else {
		fields["contractAddress"] = nil
	}

	if tx.GetTo() != nil {
		fields["to"] = common.BytesToAddress(tx.GetTo().Bytes())
	}

	api.logger.Debugf("eth_getTransactionReceipt: %v", fields)

	return fields, nil
}

// SendRawTransaction send a raw Ethereum transaction.
func (api *TransactionAPI) SendRawTransaction(data hexutil.Bytes) (ret common.Hash, err error) {
	defer func(start time.Time) {
		invokeSendRawTxDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	if api.rep.ReadonlyMode {
		return [32]byte{}, errors.New("readonly mode cannot process tx")
	}

	tx := &types.Transaction{}
	if err := tx.Unmarshal(data); err != nil {
		return [32]byte{}, err
	}

	if api.rep.Config.Access.EnableWhiteList {
		from := tx.GetFrom()
		if from == nil {
			return [32]byte{}, errors.New("verify tx err")
		}
		stateLedger, err := getStateLedgerAt(api.api, nil) // use the latest block
		if err != nil {
			return [32]byte{}, err
		}
		if err := access.Verify(stateLedger, from.String()); err != nil {
			return [32]byte{}, err
		}
	}

	if err := checkTransaction(tx); err != nil {
		return [32]byte{}, fmt.Errorf("check transaction fail for %s", err.Error())
	}

	if err := api.api.Broker().ConsensusReady(); err != nil {
		if api.rep.Config.JsonRPC.RejectTxsIfConsensusAbnormal {
			return [32]byte{}, fmt.Errorf("the system is temporarily unavailable %s, tx: %s", err.Error(), tx.GetHash().String())
		}
	} else {
		api.logger.Debugf("Receive new eth tx: %s", tx.GetHash().String())
	}

	return sendTransaction(api.api, tx)
}

func getTxByBlockInfoAndIndex(api api.CoreAPI, mode string, key string, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	block, err := api.Broker().GetBlock(mode, key)
	if err != nil {
		return nil, err
	}

	if int(idx) >= len(block.Transactions) {
		return nil, errors.New("index beyond block transactions' size")
	}

	tx := block.Transactions[idx]

	meta, err := api.Broker().GetTransactionMeta(tx.GetHash())
	if err != nil {
		return nil, err
	}

	return NewRPCTransaction(tx, common.BytesToHash(meta.BlockHash.Bytes()), meta.BlockHeight, meta.Index), nil
}

func checkTransaction(tx *types.Transaction) error {
	if tx.GetTo() == nil {
		if len(tx.GetPayload()) == 0 {
			return errors.New("can't deploy empty contract")
		}
	}

	if len(tx.GetSignature()) == 0 {
		return errors.New("signature can't be empty")
	}

	return nil
}

func sendTransaction(api api.CoreAPI, tx *types.Transaction) (common.Hash, error) {
	err := api.Broker().HandleTransaction(tx)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.GetHash().ETHHash(), nil
}
