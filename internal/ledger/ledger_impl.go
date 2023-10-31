package ledger

import (
	"fmt"
	"time"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/blockfile"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type Ledger struct {
	ChainLedger ChainLedger
	StateLedger StateLedger
}

type BlockData struct {
	Block      *types.Block
	Receipts   []*types.Receipt
	Accounts   map[string]IAccount
	TxHashList []*types.Hash
}

func NewLedgerWithStores(repo *repo.Repo, blockchainStore storage.Storage, ldb storage.Storage, bf *blockfile.BlockFile) (*Ledger, error) {
	var err error
	ledger := &Ledger{}
	if blockchainStore != nil || bf != nil {
		ledger.ChainLedger, err = newChainLedger(repo, blockchainStore, bf)
		if err != nil {
			return nil, fmt.Errorf("init chain ledger failed: %w", err)
		}
	} else {
		ledger.ChainLedger, err = NewChainLedger(repo, "")
		if err != nil {
			return nil, fmt.Errorf("init chain ledger failed: %w", err)
		}
	}
	if ldb != nil {
		ledger.StateLedger, err = newStateLedger(repo, ldb)
		if err != nil {
			return nil, fmt.Errorf("init state ledger failed: %w", err)
		}
	} else {
		ledger.StateLedger, err = NewStateLedger(repo, "")
		if err != nil {
			return nil, fmt.Errorf("init state ledger failed: %w", err)
		}
	}

	meta := ledger.ChainLedger.GetChainMeta()
	if err := ledger.Rollback(meta.Height); err != nil {
		return nil, fmt.Errorf("rollback ledger to height %d failed: %w", meta.Height, err)
	}

	return ledger, nil
}

func NewLedger(rep *repo.Repo) (*Ledger, error) {
	return NewLedgerWithStores(rep, nil, nil, nil)
}

// PersistBlockData persists block data
func (l *Ledger) PersistBlockData(blockData *BlockData) {
	current := time.Now()
	block := blockData.Block
	receipts := blockData.Receipts
	accounts := blockData.Accounts

	err := l.StateLedger.Commit(block.BlockHeader.Number, accounts, block.BlockHeader.StateRoot)
	if err != nil {
		panic(err)
	}
	flushDirtyWorldStateDuration.Observe(float64(time.Since(current)) / float64(time.Second))

	current = time.Now()
	if err := l.ChainLedger.PersistExecutionResult(block, receipts); err != nil {
		panic(err)
	}

	persistBlockDuration.Observe(float64(time.Since(current)) / float64(time.Second))
	blockHeightMetric.Set(float64(block.BlockHeader.Number))
}

// Rollback rollback ledger to history version
func (l *Ledger) Rollback(height uint64) error {
	if err := l.StateLedger.RollbackState(height); err != nil {
		return fmt.Errorf("rollback state to height %d failed: %w", height, err)
	}

	if err := l.ChainLedger.RollbackBlockChain(height); err != nil {
		return fmt.Errorf("rollback block to height %d failed: %w", height, err)
	}

	blockHeightMetric.Set(float64(height))
	return nil
}

func (l *Ledger) Close() {
	l.ChainLedger.Close()
	l.StateLedger.Close()
}

func (l *Ledger) NewView() *Ledger {
	return &Ledger{
		ChainLedger: l.ChainLedger,
		StateLedger: l.StateLedger.NewView(),
	}
}
