package ledger

import (
	"errors"
	"fmt"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/snapshot"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var _ StateLedger = (*StateLedgerImpl)(nil)

var (
	ErrorRollbackToHigherNumber = errors.New("rollback to higher blockchain height")
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
	triePreloader *triePreloader
	accounts      map[string]IAccount
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

	snapshotDB      storage.Storage
	snapshotCacheDB storage.Storage
	snapshot        *snapshot.Snapshot

	transientStorage transientStorage

	// enableExpensiveMetric determines if costly metrics gathering is allowed or not.
	// The goal is to separate standard metrics for health monitoring and debug metrics that might impact runtime performance.
	enableExpensiveMetric bool
}

// NewView get a view at specific block. We can enable snapshot if and only if the block were the latest block.
func (l *StateLedgerImpl) NewView(block *types.Block, enableSnapshot bool) StateLedger {
	l.logger.Debugf("[NewView] height: %v, stateRoot: %v", block.BlockHeader.Number, block.BlockHeader.StateRoot)
	// TODO(zqr): multi snapshot layers can also support view ledger
	lg := &StateLedgerImpl{
		repo:                  l.repo,
		logger:                l.logger,
		db:                    l.db,
		cachedDB:              l.cachedDB,
		accountCache:          l.accountCache,
		accounts:              make(map[string]IAccount),
		preimages:             make(map[types.Hash][]byte),
		changer:               NewChanger(),
		accessList:            NewAccessList(),
		logs:                  NewEvmLogs(),
		enableExpensiveMetric: l.enableExpensiveMetric,
	}
	if enableSnapshot {
		lg.snapshotDB = l.snapshotDB
		lg.snapshotCacheDB = l.snapshotCacheDB
		lg.snapshot = l.snapshot
	}
	lg.refreshAccountTrie(block.BlockHeader.StateRoot)
	return lg
}

// NewViewWithoutCache get a view ledger at specific block. We can enable snapshot if and only if the block were the latest block.
func (l *StateLedgerImpl) NewViewWithoutCache(block *types.Block, enableSnapshot bool) StateLedger {
	l.logger.Debugf("[NewViewWithoutCache] height: %v, stateRoot: %v", block.BlockHeader.Number, block.BlockHeader.StateRoot)
	ac, _ := NewAccountCache(0, true)
	// TODO(zqr): multi snapshot layers can also support historical view ledger
	lg := &StateLedgerImpl{
		repo:                  l.repo,
		logger:                l.logger,
		db:                    l.db,
		cachedDB:              l.db,
		accountCache:          ac,
		accounts:              make(map[string]IAccount),
		preimages:             make(map[types.Hash][]byte),
		changer:               NewChanger(),
		accessList:            NewAccessList(),
		logs:                  NewEvmLogs(),
		enableExpensiveMetric: l.enableExpensiveMetric,
	}
	if enableSnapshot {
		lg.snapshotDB = l.snapshotDB
		lg.snapshotCacheDB = l.snapshotCacheDB
		lg.snapshot = l.snapshot
	}
	lg.refreshAccountTrie(block.BlockHeader.StateRoot)
	return lg
}

func (l *StateLedgerImpl) Finalise() {
	for _, account := range l.accounts {
		keys := account.Finalise()

		if l.triePreloader != nil {
			l.triePreloader.preload(common.Hash{}, [][]byte{compositeAccountKey(account.GetAddress())})
			if len(keys) > 0 {
				l.triePreloader.preload(account.GetStorageRootHash(), keys)
			}
		}
	}

	l.ClearChangerAndRefund()
}

func newStateLedger(rep *repo.Repo, stateStorage, snapshotStorage storage.Storage) (StateLedger, error) {
	cachedStateStorage := storagemgr.NewCachedStorage(stateStorage, rep.Config.Ledger.StateLedgerCacheMegabytesLimit)

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
		accounts:              make(map[string]IAccount),
		preimages:             make(map[types.Hash][]byte),
		changer:               NewChanger(),
		accessList:            NewAccessList(),
		logs:                  NewEvmLogs(),
		enableExpensiveMetric: rep.Config.Monitor.EnableExpensive,
	}

	if snapshotStorage != nil {
		snapshotCachedStorage := storagemgr.NewCachedStorage(snapshotStorage, rep.Config.Snapshot.DiskCacheMegabytesLimit)
		ledger.snapshotDB = snapshotStorage
		ledger.snapshotCacheDB = snapshotCachedStorage
		ledger.snapshot = snapshot.NewSnapshot(snapshotCachedStorage, ledger.logger)
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

	snapshotStoragePath := repo.GetStoragePath(rep.RepoRoot, storagemgr.Snapshot)
	if storageDir != "" {
		snapshotStoragePath = path.Join(storageDir, storagemgr.Snapshot)
	}
	snapshotStorage, err := storagemgr.Open(snapshotStoragePath)
	if err != nil {
		return nil, fmt.Errorf("create snapshot storage: %w", err)
	}

	return newStateLedger(rep, stateStorage, snapshotStorage)
}

func (l *StateLedgerImpl) SetTxContext(thash *types.Hash, ti int) {
	l.thash = thash
	l.txIndex = ti
}

// Close close the ledger instance
func (l *StateLedgerImpl) Close() {
	_ = l.cachedDB.Close()
	l.triePreloader.close()
}
