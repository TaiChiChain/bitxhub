package ledger

import (
	"errors"
	"fmt"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var _ StateLedger = (*StateLedgerImpl)(nil)

var (
	ErrorRollbackToHigherNumber  = errors.New("rollback to higher blockchain height")
	ErrorRollbackTooMuch         = errors.New("rollback too much block")
	ErrorRemoveJournalOutOfRange = errors.New("remove journal out of range")
)

type revision struct {
	id           int
	changerIndex int
}

type StateLedgerImpl struct {
	logger        logrus.FieldLogger
	db            storage.Storage
	cachedDB      storage.Storage
	accountCache  *AccountCache
	accountTrie   *jmt.JMT // keep track of the latest world state (dirty or committed)
	minJnlHeight  uint64
	maxJnlHeight  uint64
	accounts      map[string]IAccount
	blockJournals map[string]*BlockJournal
	prevJnlHash   *types.Hash
	repo          *repo.Repo
	blockHeight   uint64
	thash         *types.Hash
	txIndex       int

	validRevisions []revision
	nextRevisionId int
	changer        *stateChanger

	accessList *AccessList
	preimages  map[types.Hash][]byte
	refund     uint64
	logs       *evmLogs

	transientStorage transientStorage

	// enableExpensiveMetric determines if costly metrics gathering is allowed or not.
	// The goal is to separate standard metrics for health monitoring and debug metrics that might impact runtime performance.
	enableExpensiveMetric bool
}

// NewView get a view
func (l *StateLedgerImpl) NewView(block *types.Block) StateLedger {
	l.logger.Debugf("[NewView] height: %v, stateRoot: %v", block.BlockHeader.Number, block.BlockHeader.StateRoot)
	lg := &StateLedgerImpl{
		repo:          l.repo,
		logger:        l.logger,
		db:            l.db,
		cachedDB:      l.cachedDB,
		accountCache:  l.accountCache,
		minJnlHeight:  l.minJnlHeight,
		maxJnlHeight:  l.maxJnlHeight,
		accounts:      make(map[string]IAccount),
		prevJnlHash:   l.prevJnlHash,
		preimages:     make(map[types.Hash][]byte),
		changer:       NewChanger(),
		accessList:    NewAccessList(),
		logs:          NewEvmLogs(),
		blockJournals: make(map[string]*BlockJournal),
	}
	lg.refreshAccountTrie(block.BlockHeader.StateRoot)
	return lg
}

// NewView get a view
func (l *StateLedgerImpl) NewViewWithoutCache(blockHeader *types.BlockHeader) StateLedger {
	l.logger.Debugf("[NewViewWithoutCache] height: %v, stateRoot: %v", blockHeader.Number, blockHeader.StateRoot)
	ac, _ := NewAccountCache(0, true)
	lg := &StateLedgerImpl{
		repo:          l.repo,
		logger:        l.logger,
		db:            l.db,
		cachedDB:      l.db,
		accountCache:  ac,
		minJnlHeight:  l.minJnlHeight,
		maxJnlHeight:  l.maxJnlHeight,
		accounts:      make(map[string]IAccount),
		prevJnlHash:   l.prevJnlHash,
		preimages:     make(map[types.Hash][]byte),
		changer:       NewChanger(),
		accessList:    NewAccessList(),
		logs:          NewEvmLogs(),
		blockJournals: make(map[string]*BlockJournal),
	}
	lg.refreshAccountTrie(blockHeader.StateRoot)
	return lg
}

func (l *StateLedgerImpl) Finalise() {
	for _, account := range l.accounts {
		account.Finalise()
	}
	l.ClearChangerAndRefund()
}

func newStateLedger(rep *repo.Repo, stateStorage storage.Storage) (StateLedger, error) {
	minJnlHeight, maxJnlHeight := getJournalRange(stateStorage)
	prevJnlHash := &types.Hash{}
	if maxJnlHeight != 0 {
		blockJournal := getBlockJournal(maxJnlHeight, stateStorage)
		if blockJournal == nil {
			return nil, fmt.Errorf("get empty block journal for block: %d", maxJnlHeight)
		}
		prevJnlHash = blockJournal.ChangedHash
	}

	cachedStateStorage, err := storagemgr.NewCachedStorage(stateStorage, rep.Config.Ledger.StateLedgerCacheMegabytesLimit)
	if err != nil {
		return nil, err
	}

	accountCache, err := NewAccountCache(rep.Config.Ledger.StateLedgerAccountCacheSize, false)
	if err != nil {
		return nil, err
	}
	accountCache.SetEnableExpensiveMetric(rep.Config.Monitor.EnableExpensive)

	ledger := &StateLedgerImpl{
		repo:                  rep,
		logger:                loggers.Logger(loggers.Storage),
		db:                    stateStorage,
		cachedDB:              cachedStateStorage,
		accountCache:          accountCache,
		minJnlHeight:          minJnlHeight,
		maxJnlHeight:          maxJnlHeight,
		accounts:              make(map[string]IAccount),
		prevJnlHash:           prevJnlHash,
		preimages:             make(map[types.Hash][]byte),
		changer:               NewChanger(),
		accessList:            NewAccessList(),
		logs:                  NewEvmLogs(),
		blockJournals:         make(map[string]*BlockJournal),
		enableExpensiveMetric: rep.Config.Monitor.EnableExpensive,
	}
	ledger.refreshAccountTrie(nil)

	return ledger, nil
}

// NewStateLedger create a new ledger instance
func NewStateLedger(rep *repo.Repo, storageDir string) (StateLedger, error) {
	stateStoragePath := repo.GetStoragePath(rep.RepoRoot, storagemgr.Ledger)
	if storageDir != "" {
		stateStoragePath = path.Join(storageDir, storagemgr.Ledger)
	}
	stateStorage, err := storagemgr.Open(stateStoragePath)
	if err != nil {
		return nil, fmt.Errorf("create stateDB: %w", err)
	}

	return newStateLedger(rep, stateStorage)
}

func (l *StateLedgerImpl) SetTxContext(thash *types.Hash, ti int) {
	l.thash = thash
	l.txIndex = ti
}

// removeJournalsBeforeBlock removes ledger journals whose block number < height
func (l *StateLedgerImpl) removeJournalsBeforeBlock(height uint64) error {
	if height > l.maxJnlHeight {
		return ErrorRemoveJournalOutOfRange
	}

	if height <= l.minJnlHeight {
		return nil
	}

	batch := l.cachedDB.NewBatch()
	for i := l.minJnlHeight; i < height; i++ {
		batch.Delete(compositeKey(journalKey, i))
	}
	batch.Put(compositeKey(journalKey, minHeightStr), marshalHeight(height))
	batch.Commit()

	l.minJnlHeight = height

	return nil
}

// Close close the ledger instance
func (l *StateLedgerImpl) Close() {
	_ = l.cachedDB.Close()
}
