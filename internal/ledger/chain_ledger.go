package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path"
	"strconv"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/blockfile"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrNotFound = errors.New("not found in DB")
)

var _ ChainLedger = (*ChainLedgerImpl)(nil)

type ChainLedgerImpl struct {
	blockchainStore storage.Storage
	bf              *blockfile.BlockFile
	repo            *repo.Repo
	chainMeta       *types.ChainMeta
	logger          logrus.FieldLogger

	txCache      *lru.Cache[uint64, []*types.Transaction]
	receiptCache *lru.Cache[uint64, []*types.Receipt]
	blockCache   *lru.Cache[uint64, *types.Block]
}

func newChainLedger(rep *repo.Repo, bcStorage storage.Storage, bf *blockfile.BlockFile) (*ChainLedgerImpl, error) {
	c := &ChainLedgerImpl{
		blockchainStore: bcStorage,
		bf:              bf,
		repo:            rep,
		logger:          loggers.Logger(loggers.Storage),
	}

	var err error
	c.chainMeta, err = c.LoadChainMeta()
	if err != nil {
		return nil, fmt.Errorf("load chain meta: %w", err)
	}

	err = c.checkChainMeta()
	if err != nil {
		return nil, fmt.Errorf("check chain meta: %w", err)
	}

	txCache, err := lru.New[uint64, []*types.Transaction](rep.Config.Ledger.ChainLedgerCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new tx cache: %w", err)
	}

	receiptCache, err := lru.New[uint64, []*types.Receipt](rep.Config.Ledger.ChainLedgerCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new receipt cache: %w", err)
	}

	blockCache, err := lru.New[uint64, *types.Block](rep.Config.Ledger.ChainLedgerCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new block cache: %w", err)
	}

	c.txCache = txCache
	c.receiptCache = receiptCache
	c.blockCache = blockCache

	return c, nil
}

func NewChainLedger(rep *repo.Repo, storageDir string) (*ChainLedgerImpl, error) {
	bcStoragePath := repo.GetStoragePath(rep.RepoRoot, storagemgr.BlockChain)
	if storageDir != "" {
		bcStoragePath = path.Join(storageDir, storagemgr.BlockChain)
	}
	bcStorage, err := storagemgr.Open(bcStoragePath)
	if err != nil {
		return nil, fmt.Errorf("create blockchain storage: %w", err)
	}

	bfStoragePath := repo.GetStoragePath(rep.RepoRoot, storagemgr.Blockfile)
	if storageDir != "" {
		bfStoragePath = path.Join(storageDir, storagemgr.Blockfile)
	}
	bf, err := blockfile.NewBlockFile(bfStoragePath, loggers.Logger(loggers.Storage))
	if err != nil {
		return nil, fmt.Errorf("blockfile initialize: %w", err)
	}

	return newChainLedger(rep, bcStorage, bf)
}

// GetBlock get block with height
func (l *ChainLedgerImpl) GetBlock(height uint64) (*types.Block, error) {
	if height > l.chainMeta.Height {
		return nil, ErrNotFound
	}

	if block, ok := l.blockCache.Get(height); ok {
		return block, nil
	}
	data, err := l.bf.Get(blockfile.BlockFileBodiesTable, height)
	if err != nil {
		return nil, fmt.Errorf("get bodies with height %d from blockfile failed: %w", height, err)
	}

	block := &types.Block{}
	if err := block.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("unmarshal block error: %w", err)
	}

	txHashesData := l.blockchainStore.Get(compositeKey(blockTxSetKey, height))
	if txHashesData == nil {
		return nil, errors.New("cannot get tx hashes of block")
	}
	txHashes := make([]*types.Hash, 0)
	if err := json.Unmarshal(txHashesData, &txHashes); err != nil {
		return nil, fmt.Errorf("unmarshal tx hash data error: %w", err)
	}

	var txs []*types.Transaction
	txs, ok := l.txCache.Get(height)
	if !ok {
		txsBytes, err := l.bf.Get(blockfile.BlockFileTXsTable, height)
		if err != nil {
			return nil, fmt.Errorf("get transactions with height %d from blockfile failed: %w", height, err)
		}
		txs, err = types.UnmarshalTransactions(txsBytes)
		if err != nil {
			return nil, fmt.Errorf("unmarshal txs bytes error: %w", err)
		}
		l.txCache.Add(height, txs)
	}

	block.Transactions = txs
	l.blockCache.Add(height, block)

	return block, nil
}

func (l *ChainLedgerImpl) GetBlockHash(height uint64) *types.Hash {
	if height > l.chainMeta.Height {
		return &types.Hash{}
	}

	hash := l.blockchainStore.Get(compositeKey(blockHeightKey, height))
	if hash == nil {
		return &types.Hash{}
	}
	return types.NewHashByStr(string(hash))
}

// GetBlockSign get the signature of block
func (l *ChainLedgerImpl) GetBlockSign(height uint64) ([]byte, error) {
	block, err := l.GetBlock(height)
	if err != nil {
		return nil, fmt.Errorf("get block with height %d failed: %w", height, err)
	}

	return block.Signature, nil
}

// GetBlockByHash get the block using block hash
func (l *ChainLedgerImpl) GetBlockByHash(hash *types.Hash) (*types.Block, error) {
	data := l.blockchainStore.Get(compositeKey(blockHashKey, hash.String()))
	if data == nil {
		return nil, ErrNotFound
	}

	height, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, fmt.Errorf("wrong height, %w", err)
	}

	return l.GetBlock(uint64(height))
}

// GetTransaction get the transaction using transaction hash
func (l *ChainLedgerImpl) GetTransaction(hash *types.Hash) (*types.Transaction, error) {
	metaBytes := l.blockchainStore.Get(compositeKey(transactionMetaKey, hash.String()))
	if metaBytes == nil {
		return nil, ErrNotFound
	}
	meta := &types.TransactionMeta{}
	if err := meta.Unmarshal(metaBytes); err != nil {
		return nil, fmt.Errorf("unmarshal transaction meta bytes error: %w", err)
	}

	var txs []*types.Transaction
	txs, ok := l.txCache.Get(meta.BlockHeight)
	if !ok {
		txsBytes, err := l.bf.Get(blockfile.BlockFileTXsTable, meta.BlockHeight)
		if err != nil {
			return nil, fmt.Errorf("get transactions with height %d from blockfile failed: %w", meta.BlockHeight, err)
		}

		return types.UnmarshalTransactionWithIndex(txsBytes, meta.Index)
	}

	return txs[meta.Index], nil
}

func (l *ChainLedgerImpl) GetTransactionCount(height uint64) (uint64, error) {
	txHashesData := l.blockchainStore.Get(compositeKey(blockTxSetKey, height))
	if txHashesData == nil {
		return 0, errors.New("cannot get tx hashes of block")
	}
	txHashes := make([]types.Hash, 0)
	if err := json.Unmarshal(txHashesData, &txHashes); err != nil {
		return 0, fmt.Errorf("unmarshal tx hash data error: %w", err)
	}

	return uint64(len(txHashes)), nil
}

// GetTransactionMeta get the transaction meta data
func (l *ChainLedgerImpl) GetTransactionMeta(hash *types.Hash) (*types.TransactionMeta, error) {
	data := l.blockchainStore.Get(compositeKey(transactionMetaKey, hash.String()))
	if data == nil {
		return nil, ErrNotFound
	}

	meta := &types.TransactionMeta{}
	if err := meta.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("unmarshal transaction meta error: %w", err)
	}

	return meta, nil
}

// GetReceipt get the transaction receipt
func (l *ChainLedgerImpl) GetReceipt(hash *types.Hash) (*types.Receipt, error) {
	metaBytes := l.blockchainStore.Get(compositeKey(transactionMetaKey, hash.String()))
	if metaBytes == nil {
		return nil, ErrNotFound
	}
	meta := &types.TransactionMeta{}
	if err := meta.Unmarshal(metaBytes); err != nil {
		return nil, fmt.Errorf("unmarshal transaction meta bytes error: %w", err)
	}

	var rs []*types.Receipt
	rs, ok := l.receiptCache.Get(meta.BlockHeight)
	if !ok {
		rsBytes, err := l.bf.Get(blockfile.BlockFileReceiptTable, meta.BlockHeight)
		if err != nil {
			return nil, fmt.Errorf("get receipts with height %d from blockfile failed: %w", meta.BlockHeight, err)
		}

		return types.UnmarshalReceiptWithIndex(rsBytes, meta.Index)
	}

	return rs[meta.Index], nil
}

// PersistExecutionResult persist the execution result
func (l *ChainLedgerImpl) PersistExecutionResult(block *types.Block, receipts []*types.Receipt) error {
	current := time.Now()

	if block == nil {
		return errors.New("empty block data")
	}

	batcher := l.blockchainStore.NewBatch()

	rs, err := l.prepareReceipts(batcher, block, receipts)
	if err != nil {
		return fmt.Errorf("preapare receipts failed: %w", err)
	}

	ts, err := l.prepareTransactions(batcher, block)
	if err != nil {
		return fmt.Errorf("prepare transactions failed: %w", err)
	}

	b, err := l.prepareBlock(batcher, block)
	if err != nil {
		return fmt.Errorf("prepare block failed: %w", err)
	}

	meta := &types.ChainMeta{
		Height:    block.BlockHeader.Number,
		GasPrice:  big.NewInt(block.BlockHeader.GasPrice),
		BlockHash: block.BlockHash,
	}

	l.logger.WithFields(logrus.Fields{
		"Height":    meta.Height,
		"GasPrice":  meta.GasPrice,
		"BlockHash": meta.BlockHash,
	}).Debug("prepare chain meta")

	if err := l.bf.AppendBlock(l.chainMeta.Height, block.BlockHash.Bytes(), b, rs, ts); err != nil {
		return fmt.Errorf("append block with height %d to blockfile failed: %w", l.chainMeta.Height, err)
	}

	if err := l.persistChainMeta(batcher, meta); err != nil {
		return fmt.Errorf("persist chain meta failed: %w", err)
	}

	batcher.Commit()

	if len(block.Transactions) > 0 {
		l.txCache.Add(meta.Height, block.Transactions)
	}
	if len(receipts) > 0 {
		l.receiptCache.Add(meta.Height, receipts)
	}

	l.blockCache.Add(meta.Height, block)

	l.UpdateChainMeta(meta)

	l.logger.WithField("time", time.Since(current)).Debug("persist execution result elapsed")

	return nil
}

// UpdateChainMeta update the chain meta data
func (l *ChainLedgerImpl) UpdateChainMeta(meta *types.ChainMeta) {
	l.chainMeta.Height = meta.Height
	l.chainMeta.GasPrice = meta.GasPrice
	l.chainMeta.BlockHash = meta.BlockHash
}

// GetChainMeta get chain meta data
func (l *ChainLedgerImpl) GetChainMeta() *types.ChainMeta {
	return &types.ChainMeta{
		Height:    l.chainMeta.Height,
		GasPrice:  l.chainMeta.GasPrice,
		BlockHash: l.chainMeta.BlockHash,
	}
}

// LoadChainMeta load chain meta data
func (l *ChainLedgerImpl) LoadChainMeta() (*types.ChainMeta, error) {
	ok := l.blockchainStore.Has([]byte(chainMetaKey))

	chain := &types.ChainMeta{
		Height:    0,
		BlockHash: &types.Hash{},
	}
	if ok {
		body := l.blockchainStore.Get([]byte(chainMetaKey))
		if err := chain.Unmarshal(body); err != nil {
			return nil, fmt.Errorf("unmarshal chain meta: %w", err)
		}
	}

	return chain, nil
}

func (l *ChainLedgerImpl) prepareReceipts(_ storage.Batch, _ *types.Block, receipts []*types.Receipt) ([]byte, error) {
	return types.MarshalReceipts(receipts)
}

func (l *ChainLedgerImpl) prepareTransactions(batcher storage.Batch, block *types.Block) ([]byte, error) {
	for i, tx := range block.Transactions {
		meta := &types.TransactionMeta{
			BlockHeight: block.BlockHeader.Number,
			BlockHash:   block.BlockHash,
			Index:       uint64(i),
		}

		metaBytes, err := meta.Marshal()
		if err != nil {
			return nil, fmt.Errorf("marshal tx meta error: %s", err)
		}

		batcher.Put(compositeKey(transactionMetaKey, tx.GetHash().String()), metaBytes)
	}

	return types.MarshalTransactions(block.Transactions)
}

func (l *ChainLedgerImpl) prepareBlock(batcher storage.Batch, block *types.Block) ([]byte, error) {
	// Generate block header signature
	if block.Signature == nil {
		signed, err := l.repo.AccountKeySign(block.BlockHash.Bytes())
		if err != nil {
			return nil, fmt.Errorf("sign block %s failed: %w", block.BlockHash.String(), err)
		}

		block.Signature = signed
	}

	storedBlock := &types.Block{
		BlockHeader:  block.BlockHeader,
		Transactions: nil,
		BlockHash:    block.BlockHash,
		Signature:    block.Signature,
		Extra:        block.Extra,
	}
	bs, err := storedBlock.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal stored block error: %w", err)
	}

	height := block.BlockHeader.Number

	var txHashes []*types.Hash
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.GetHash())
	}

	data, err := json.Marshal(txHashes)
	if err != nil {
		return nil, fmt.Errorf("marshal tx hash error: %w", err)
	}

	batcher.Put(compositeKey(blockTxSetKey, height), data)

	hash := block.BlockHash.String()
	batcher.Put(compositeKey(blockHashKey, hash), []byte(fmt.Sprintf("%d", height)))
	batcher.Put(compositeKey(blockHeightKey, height), []byte(hash))

	return bs, nil
}

func (l *ChainLedgerImpl) persistChainMeta(batcher storage.Batch, meta *types.ChainMeta) error {
	data, err := meta.Marshal()
	if err != nil {
		return fmt.Errorf("marshal chain meta error: %w", err)
	}

	batcher.Put([]byte(chainMetaKey), data)

	return nil
}

func (l *ChainLedgerImpl) removeChainDataOnBlock(batch storage.Batch, height uint64) error {
	block, err := l.GetBlock(height)
	if err != nil {
		return fmt.Errorf("get block with height %d failed: %w", height, err)
	}

	if err := l.bf.TruncateBlocks(height - 1); err != nil {
		return fmt.Errorf("truncate blocks failed: %w", err)
	}

	batch.Delete(compositeKey(blockTxSetKey, height))
	batch.Delete(compositeKey(blockHashKey, block.BlockHash.String()))
	batch.Delete(compositeKey(interchainMetaKey, height))

	for _, tx := range block.Transactions {
		batch.Delete(compositeKey(transactionMetaKey, tx.GetHash().String()))
	}

	l.txCache.Remove(height)
	l.receiptCache.Remove(height)
	l.blockCache.Remove(height)

	return nil
}

func (l *ChainLedgerImpl) RollbackBlockChain(height uint64) error {
	meta := l.GetChainMeta()

	if meta.Height < height {
		return ErrorRollbackToHigherNumber
	}

	if meta.Height == height {
		return nil
	}

	batch := l.blockchainStore.NewBatch()

	for i := meta.Height; i > height; i-- {
		err := l.removeChainDataOnBlock(batch, i)
		if err != nil {
			return fmt.Errorf("remove chain data on block %d failed: %w", i, err)
		}
	}

	if height == 0 {
		batch.Delete([]byte(chainMetaKey))
		meta = &types.ChainMeta{}
	} else {
		block, err := l.GetBlock(height)
		if err != nil {
			return fmt.Errorf("get block with height %d failed: %w", height, err)
		}
		meta = &types.ChainMeta{
			Height:    block.BlockHeader.Number,
			GasPrice:  big.NewInt(block.BlockHeader.GasPrice),
			BlockHash: block.BlockHash,
		}

		if err := l.persistChainMeta(batch, meta); err != nil {
			return fmt.Errorf("persist chain meta failed: %w", err)
		}
	}

	batch.Commit()

	l.UpdateChainMeta(meta)

	return nil
}

func (l *ChainLedgerImpl) checkChainMeta() error {
	bfHeight, err := l.bf.Blocks()
	if err != nil {
		return fmt.Errorf("get blockfile height: %w", err)
	}
	if l.chainMeta.Height < bfHeight {
		l.chainMeta.Height = bfHeight
		batcher := l.blockchainStore.NewBatch()

		// Get block body
		data, err := l.bf.Get(blockfile.BlockFileBodiesTable, bfHeight)
		if err != nil {
			return fmt.Errorf("get bodies with height %d from blockfile failed: %w", bfHeight, err)
		}
		currentBlock := &types.Block{}
		if err := currentBlock.Unmarshal(data); err != nil {
			return fmt.Errorf("unmarshal block error: %w", err)
		}
		if err != nil {
			return fmt.Errorf("get blockfile block: %w", err)
		}

		// Get txs
		txsBytes, err := l.bf.Get(blockfile.BlockFileTXsTable, bfHeight)
		if err != nil {
			return fmt.Errorf("get transactions with height %d from blockfile failed: %w", bfHeight, err)
		}
		txs, err := types.UnmarshalTransactions(txsBytes)
		if err != nil {
			return fmt.Errorf("unmarshal txs bytes error: %w", err)
		}
		currentBlock.Transactions = txs
		_, err = l.prepareTransactions(batcher, currentBlock)
		if err != nil {
			return err
		}
		_, err = l.prepareBlock(batcher, currentBlock)
		if err != nil {
			return err
		}

		l.chainMeta.GasPrice = big.NewInt(currentBlock.BlockHeader.GasPrice)
		l.chainMeta.BlockHash = currentBlock.BlockHash
		if err := l.persistChainMeta(batcher, l.chainMeta); err != nil {
			return fmt.Errorf("update chain meta: %w", err)
		}

		batcher.Commit()
	}
	if l.chainMeta.Height > bfHeight {
		panic("illegal chain meta!")
	}

	return nil
}

func (l *ChainLedgerImpl) Close() {
	_ = l.blockchainStore.Close()
	_ = l.bf.Close()
}

func (l *ChainLedgerImpl) CloseBlockfile() {
	_ = l.bf.Close()
}
