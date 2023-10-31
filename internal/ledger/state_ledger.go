package ledger

import (
	"errors"
	"fmt"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var _ StateLedger = (*StateLedgerImpl)(nil)

var (
	ErrorRollbackToHigherNumber  = errors.New("rollback to higher blockchain height")
	ErrorRollbackWithoutJournal  = errors.New("rollback to blockchain height without journal")
	ErrorRollbackTooMuch         = errors.New("rollback too much block")
	ErrorRemoveJournalOutOfRange = errors.New("remove journal out of range")
)

type revision struct {
	id           int
	changerIndex int
}

type StateLedgerImpl struct {
	logger        logrus.FieldLogger
	ldb           storage.Storage
	minJnlHeight  uint64
	maxJnlHeight  uint64
	accounts      map[string]IAccount
	accountCache  *AccountCache
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
func (l *StateLedgerImpl) NewView() StateLedger {
	return &StateLedgerImpl{
		repo:          l.repo,
		logger:        l.logger,
		ldb:           l.ldb,
		minJnlHeight:  l.minJnlHeight,
		maxJnlHeight:  l.maxJnlHeight,
		accounts:      make(map[string]IAccount),
		accountCache:  l.accountCache,
		prevJnlHash:   l.prevJnlHash,
		preimages:     make(map[types.Hash][]byte),
		changer:       NewChanger(),
		accessList:    NewAccessList(),
		logs:          NewEvmLogs(),
		blockJournals: make(map[string]*BlockJournal),
	}
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

	accountCache, err := NewAccountCache()
	if err != nil {
		return nil, fmt.Errorf("init account cache failed: %w", err)
	}
	accountCache.enableExpensiveMetric = rep.Config.Monitor.EnableExpensive

	ledger := &StateLedgerImpl{
		repo:                  rep,
		logger:                loggers.Logger(loggers.Storage),
		ldb:                   stateStorage,
		minJnlHeight:          minJnlHeight,
		maxJnlHeight:          maxJnlHeight,
		accounts:              make(map[string]IAccount),
		accountCache:          accountCache,
		prevJnlHash:           prevJnlHash,
		preimages:             make(map[types.Hash][]byte),
		changer:               NewChanger(),
		accessList:            NewAccessList(),
		logs:                  NewEvmLogs(),
		blockJournals:         make(map[string]*BlockJournal),
		enableExpensiveMetric: rep.Config.Monitor.EnableExpensive,
	}
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

func (l *StateLedgerImpl) AccountCache() *AccountCache {
	return l.accountCache
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

	batch := l.ldb.NewBatch()
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
	_ = l.ldb.Close()
}
