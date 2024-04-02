package snapshot

import (
	"errors"
	"fmt"
	"github.com/axiomesh/axiom-kit/storage/blockjournal"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"math/big"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type Snapshot struct {
	rep *repo.Repo

	accountSnapshotCache  *storagemgr.CacheWrapper
	contractSnapshotCache *storagemgr.CacheWrapper
	backend               storage.Storage
	blockjournal          *blockjournal.BlockJournal

	lock sync.RWMutex

	logger logrus.FieldLogger
}

var (
	ErrorRollbackToHigherNumber  = errors.New("rollback snapshot to higher blockchain height")
	ErrorRollbackTooMuch         = errors.New("rollback snapshot too much block")
	ErrorRemoveJournalOutOfRange = errors.New("remove snapshot journal out of range")
)

// maxBatchSize defines the maximum size of the data in single batch write operation, which is 64 MB.
const maxBatchSize = 64 * 1024 * 1024

func NewSnapshot(rep *repo.Repo, backend storage.Storage, logger logrus.FieldLogger) *Snapshot {
	s := &Snapshot{
		rep:                   rep,
		accountSnapshotCache:  storagemgr.NewCacheWrapper(rep.Config.Snapshot.AccountSnapshotCacheMegabytesLimit, true),
		contractSnapshotCache: storagemgr.NewCacheWrapper(rep.Config.Snapshot.ContractSnapshotCacheMegabytesLimit, true),
		backend:               backend,
		logger:                logger,
	}
	bj, err := blockjournal.NewBlockJournal(repo.GetStoragePath(rep.RepoRoot, storagemgr.BlockJournal), storagemgr.BlockJournalSnapshotName, loggers.Logger(loggers.Storage))
	if err != nil {
		panic(fmt.Errorf("Create BlockJournalSnapshot err:%s", err))
	}
	s.blockjournal = bj
	return s
}

// RemoveJournalsBeforeBlock removes snapshot journals whose block number < height
func (snap *Snapshot) RemoveJournalsBeforeBlock(height uint64) error {
	minHeight, maxHeight := snap.blockjournal.GetJournalRange()
	if height > maxHeight {
		return ErrorRemoveJournalOutOfRange
	}

	if height <= minHeight {
		return nil
	}
	return snap.blockjournal.RemoveJournalsBeforeBlock(height)
}

func (snap *Snapshot) Account(addr *types.Address) (*types.InnerAccount, error) {
	snap.lock.RLock()
	defer snap.lock.RUnlock()

	accountKey := utils.CompositeAccountKey(addr)

	// try in cache first
	if snap.accountSnapshotCache != nil {
		if blob, ok := snap.accountSnapshotCache.Get(accountKey); ok {
			if len(blob) == 0 { // can be both nil and []byte{}
				return nil, nil
			}
			innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
			if err := innerAccount.Unmarshal(blob); err != nil {
				panic(err)
			}
			return innerAccount, nil
		}
	}

	// try in kv last
	blob := snap.backend.Get(accountKey)
	if snap.accountSnapshotCache != nil {
		snap.accountSnapshotCache.Set(accountKey, blob)
	}
	if len(blob) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
	innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
	if err := innerAccount.Unmarshal(blob); err != nil {
		panic(err)
	}

	return innerAccount, nil
}

// todo check correctness
func (snap *Snapshot) Storage(addr *types.Address, key []byte) ([]byte, error) {
	snap.lock.RLock()
	defer snap.lock.RUnlock()

	snapKey := utils.CompositeStorageKey(addr, key)

	if snap.contractSnapshotCache != nil {
		if blob, ok := snap.contractSnapshotCache.Get(snapKey); ok {
			return blob, nil
		}
	}

	blob := snap.backend.Get(snapKey)
	if snap.contractSnapshotCache != nil {
		snap.contractSnapshotCache.Set(snapKey, blob)
	}
	return blob, nil
}

// todo batch update snapshot of serveral blocks
func (snap *Snapshot) Update(height uint64, journal *types.SnapshotJournal, destructs map[string]struct{}, accounts map[string]*types.InnerAccount, storage map[string]map[string][]byte) error {
	snap.lock.Lock()
	defer snap.lock.Unlock()

	snap.logger.Infof("[Snapshot-Update] update snapshot at height:%v", height)

	batch := snap.backend.NewBatch()

	for addr := range destructs {
		accountKey := utils.CompositeAccountKey(types.NewAddressByStr(addr))
		batch.Delete(accountKey)
		snap.accountSnapshotCache.Del(accountKey)
	}

	for addr, acc := range accounts {
		accountKey := utils.CompositeAccountKey(types.NewAddressByStr(addr))
		blob, err := acc.Marshal()
		if err != nil {
			panic(err)
		}
		snap.accountSnapshotCache.Set(accountKey, blob)
		batch.Put(accountKey, blob)
	}

	for rawAddr, slots := range storage {
		addr := types.NewAddressByStr(rawAddr)
		for slot, blob := range slots {
			storageKey := utils.CompositeStorageKey(addr, []byte(slot))
			snap.contractSnapshotCache.Set(storageKey, blob)
			batch.Put(storageKey, blob)
		}
	}

	data, err := journal.Encode()
	if err != nil {
		return fmt.Errorf("marshal snapshot journal error: %w", err)
	}
	snap.logger.Infof("[Snapshot-Update] journal size = %v", len(data))

	batch.Commit()

	return snap.blockjournal.Append(height, data)

}

// Rollback removes snapshot journals whose block number < height
func (snap *Snapshot) Rollback(height uint64) error {
	if snap.contractSnapshotCache != nil {
		snap.contractSnapshotCache.Reset()
	}

	if snap.accountSnapshotCache != nil {
		snap.accountSnapshotCache.Reset()
	}

	minHeight, maxHeight := snap.blockjournal.GetJournalRange()
	snap.logger.Infof("[Snapshot-Rollback] minHeight=%v,maxHeight=%v,height=%v", minHeight, maxHeight, height)

	// empty snapshot, no-op
	if minHeight == 0 && maxHeight == 0 {
		return nil
	}

	if maxHeight < height {
		return ErrorRollbackToHigherNumber
	}

	if minHeight > height && !(minHeight == 1 && height == 0) {
		return ErrorRollbackTooMuch
	}

	if maxHeight == height {
		return nil
	}

	batch := snap.backend.NewBatch()
	for i := maxHeight; i > height; i-- {
		snap.logger.Infof("[Snapshot-Rollback] try executing snapshot journal of height %v", i)
		blockJournal, err := snap.GetBlockJournal(i)
		if err != nil {
			return err
		}
		if blockJournal == nil {
			snap.logger.Warnf("[Snapshot-Rollback] snapshot journal is empty at height: %v", i)
		} else {
			for _, entry := range blockJournal.Journals {
				snap.logger.Debugf("[Snapshot-Rollback] execute entry: %v", entry.String())
				revertJournal(entry, batch)
			}
		}

	}
	batch.Commit()

	err := snap.blockjournal.Truncate(height)
	if err != nil {
		return err
	}

	return nil
}

func (snap *Snapshot) ExportMetrics() {
	accountCacheMetrics := snap.accountSnapshotCache.ExportMetrics()
	snapshotAccountCacheMissCounterPerBlock.Set(float64(accountCacheMetrics.CacheMissCounter))
	snapshotAccountCacheHitCounterPerBlock.Set(float64(accountCacheMetrics.CacheHitCounter))
	snapshotAccountCacheSize.Set(float64(accountCacheMetrics.CacheSize / 1024 / 1024))
}

func (snap *Snapshot) ResetMetrics() {
	snap.accountSnapshotCache.ResetCounterMetrics()
}

func (snap *Snapshot) Batch() storage.Batch {
	return snap.backend.NewBatch()
}
