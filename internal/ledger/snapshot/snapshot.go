package snapshot

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
)

type Snapshot struct {
	snapStorage storage.Storage
	cache       *fastcache.Cache // Cache to avoid hitting the disk for direct access

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

func NewSnapshot(snapStorage storage.Storage, logger logrus.FieldLogger) *Snapshot {
	return &Snapshot{
		snapStorage: snapStorage,
		logger:      logger,
		cache:       fastcache.New(128 * 1024 * 1024), // todo configurable
	}
}

// RemoveJournalsBeforeBlock removes snapshot journals whose block number < height
func (snap *Snapshot) RemoveJournalsBeforeBlock(height uint64) error {
	minHeight, maxHeight := snap.GetJournalRange()
	if height > maxHeight {
		return ErrorRemoveJournalOutOfRange
	}

	if height <= minHeight {
		return nil
	}

	batch := snap.snapStorage.NewBatch()
	for i := minHeight; i < height; i++ {
		batch.Delete(utils.CompositeKey(utils.SnapshotKey, i))
	}
	batch.Put(utils.CompositeKey(utils.SnapshotKey, utils.MinHeightStr), utils.MarshalHeight(height))
	batch.Commit()

	return nil
}

func (snap *Snapshot) Account(addr *types.Address) (*types.InnerAccount, error) {
	snap.lock.RLock()
	defer snap.lock.RUnlock()

	accountKey := utils.CompositeAccountKey(addr)

	// Try to retrieve the account from the memory cache
	if blob, found := snap.cache.HasGet(nil, accountKey); found {
		innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
		if err := innerAccount.Unmarshal(blob); err != nil {
			panic(err)
		}
		return innerAccount, nil
	}

	blob := snap.snapStorage.Get(accountKey)
	if len(blob) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
	snap.cache.Set(accountKey, blob)

	innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
	if err := innerAccount.Unmarshal(blob); err != nil {
		panic(err)
	}

	return innerAccount, nil
}

func (snap *Snapshot) Storage(addr *types.Address, key []byte) ([]byte, error) {
	snap.lock.RLock()
	defer snap.lock.RUnlock()

	snapKey := utils.CompositeStorageKey(addr, key)

	if blob, found := snap.cache.HasGet(nil, snapKey); found {
		return blob, nil
	}
	blob := snap.snapStorage.Get(snapKey)
	if blob == nil {
		return nil, nil
	}

	snap.cache.Set(snapKey, blob)

	return blob, nil
}

func (snap *Snapshot) Update(height uint64, journal *BlockJournal, destructs map[string]struct{}, accounts map[string]*types.InnerAccount, storage map[string]map[string][]byte) error {
	snap.lock.Lock()
	defer snap.lock.Unlock()

	snap.logger.Infof("[Snapshot-Update] update snapshot at height:%v", height)

	batch := snap.snapStorage.NewBatch()

	for addr := range destructs {
		accountKey := utils.CompositeAccountKey(types.NewAddressByStr(addr))
		batch.Delete(accountKey)
		snap.cache.Del(accountKey)
	}

	for addr, acc := range accounts {
		accountKey := utils.CompositeAccountKey(types.NewAddressByStr(addr))
		blob, err := acc.Marshal()
		if err != nil {
			panic(err)
		}
		batch.Put(accountKey, blob)
		snap.cache.Set(accountKey, blob)
	}

	for rawAddr, slots := range storage {
		addr := types.NewAddressByStr(rawAddr)
		for slot, blob := range slots {
			storageKey := utils.CompositeStorageKey(addr, []byte(slot))
			batch.Put(storageKey, blob)
			snap.cache.Set(storageKey, blob)
		}
	}

	if journal != nil {
		// persist snapshot journal with meta info
		for _, entry := range journal.Journals {
			entry.PrevStates = encodeToBase64(entry.PrevStates)
		}
	}

	data, err := json.Marshal(journal)
	if err != nil {
		return fmt.Errorf("marshal snapshot journal error: %w", err)
	}

	batch.Put(utils.CompositeKey(utils.SnapshotKey, height), data)
	batch.Put(utils.CompositeKey(utils.SnapshotKey, utils.MaxHeightStr), utils.MarshalHeight(height))
	if height == 1 {
		batch.Put(utils.CompositeKey(utils.SnapshotKey, utils.MinHeightStr), utils.MarshalHeight(height))
	}
	batch.Commit()
	return nil
}

// Rollback removes snapshot journals whose block number < height
func (snap *Snapshot) Rollback(height uint64) error {
	minHeight, maxHeight := snap.GetJournalRange()
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

	batch := snap.snapStorage.NewBatch()
	for i := maxHeight; i > height; i-- {
		snap.logger.Infof("[Snapshot-Rollback] execute snapshot journal of height %v", i)
		blockJournal := snap.GetBlockJournal(i)
		if blockJournal == nil {
			return ErrorRemoveJournalOutOfRange
		}
		for _, entry := range blockJournal.Journals {
			snap.logger.Debugf("[Snapshot-Rollback] execute entry: %v", entry.String())
			revertJournal(entry, batch)
		}
		batch.Delete(utils.CompositeKey(utils.SnapshotKey, i))
		batch.Put(utils.CompositeKey(utils.SnapshotKey, utils.MaxHeightStr), utils.MarshalHeight(i-1))
		if batch.Size() > maxBatchSize {
			batch.Commit()
			batch.Reset()
			snap.logger.Infof("[Snapshot-Rollback] write batch periodically")
		}
	}
	batch.Commit()

	snap.cache.Reset()
	return nil
}

func (snap *Snapshot) Batch() storage.Batch {
	return snap.snapStorage.NewBatch()
}

func encodeToBase64(src map[string][]byte) map[string][]byte {
	encodedData := make(map[string][]byte)
	for key, value := range src {
		encodedKey := base64.StdEncoding.EncodeToString([]byte(key))
		encodedData[encodedKey] = value
	}
	return encodedData
}
