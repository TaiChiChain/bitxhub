package snapshot

import (
	"math/big"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

type diskLayer struct {
	diskdb storage.Storage
	cache  *fastcache.Cache // Cache to avoid hitting the disk for direct access

	stateRoot common.Hash // Block StateRoot hash associated with the base snapshot
	available bool

	lock sync.RWMutex
}

func NewDiskLayer(db storage.Storage) Layer {
	return &diskLayer{
		diskdb:    db,
		cache:     fastcache.New(128 * 1024 * 1024), // todo configurable
		available: true,
	}
}

// Root returns stateRoot hash for which this snapshot was made.
func (dl *diskLayer) Root() common.Hash {
	return dl.stateRoot
}

// Parent always returns nil as there's no layer below the disk.
func (dl *diskLayer) Parent() Layer {
	return nil
}

func (dl *diskLayer) Available() bool {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return dl.available
}

func (dl *diskLayer) Update(stateRoot common.Hash, destructs map[string]struct{}, accounts map[string]*types.InnerAccount, storage map[string]map[string][]byte) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	batch := dl.diskdb.NewBatch()

	for addr := range destructs {
		accountKey := CompositeSnapAccountKey(addr)
		batch.Delete(accountKey)
		dl.cache.Set(accountKey, nil)
	}

	for addr, acc := range accounts {
		accountKey := CompositeSnapAccountKey(addr)
		blob, err := acc.Marshal()
		if err != nil {
			panic(err)
		}
		batch.Put(accountKey, blob)
		dl.cache.Set(accountKey, blob)
	}

	for addr, slots := range storage {
		for slot, blob := range slots {
			storageKey := CompositeSnapStorageKey(addr, []byte(slot))
			batch.Put(storageKey, blob)
			dl.cache.Set(storageKey, blob)
		}
	}

	batch.Commit()
	dl.stateRoot = stateRoot
}

func (dl *diskLayer) Account(addr *types.Address) (*types.InnerAccount, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	if !dl.available {
		return nil, ErrSnapshotUnavailable
	}

	accountKey := CompositeSnapAccountKey(addr.String())

	// Try to retrieve the account from the memory cache
	if blob, found := dl.cache.HasGet(nil, accountKey); found {
		innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
		if err := innerAccount.Unmarshal(blob); err != nil {
			panic(err)
		}
		return innerAccount, nil
	}

	blob := dl.diskdb.Get(accountKey)
	if len(blob) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
	dl.cache.Set(accountKey, blob)

	innerAccount := &types.InnerAccount{Balance: big.NewInt(0)}
	if err := innerAccount.Unmarshal(blob); err != nil {
		panic(err)
	}

	return innerAccount, nil
}

func (dl *diskLayer) Storage(addr *types.Address, key []byte) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	if !dl.available {
		return nil, ErrSnapshotUnavailable
	}

	snapKey := CompositeSnapStorageKey(addr.String(), key)

	if blob, found := dl.cache.HasGet(nil, snapKey); found {
		return blob, nil
	}
	blob := dl.diskdb.Get(snapKey)
	if blob == nil {
		return nil, nil
	}

	dl.cache.Set(snapKey, blob)

	return blob, nil
}

func (dl *diskLayer) Clear() {
	dl.cache.Reset()
}
