package prune

import (
	"errors"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

// PruneCache enables trie node caches, so that every trie read op will happen in trie cache first,
// which avoids frequent reading from disk.
// By using PruneCache, we also implement trie pruning schema, which reduces state storage size that a full node must hold.
type PruneCache struct {
	rep *repo.Repo
	// todo check content of cache, ensure there are only trie nodes
	ledgerStorage storage.Storage
	states        *states

	prunner *prunner

	logger logrus.FieldLogger
}

type states struct {
	size      atomic.Int32
	diffs     []*diffLayer
	lock      sync.RWMutex
	allKeyMap map[string]struct{}
}

func (s *states) rebuildAllKeyMap() {
	s.allKeyMap = make(map[string]struct{})
	if len(s.diffs) > 0 {
		for _, diff := range s.diffs {
			for k, _ := range diff.accountCache {
				s.allKeyMap[k] = struct{}{}
			}
			for k, _ := range diff.storageCache {
				s.allKeyMap[k] = struct{}{}
			}
		}
	}
}

var (
	ErrorRollbackToHigherNumber = errors.New("rollback PruneCache to higher blockchain height")
	ErrorRollbackTooMuch        = errors.New("rollback PruneCache too much block")
)

func NewTrieCache(rep *repo.Repo, ledgerStorage storage.Storage, accountTrieStorage *storagemgr.CachedStorage, storageTrieStorage *storagemgr.CachedStorage, logger logrus.FieldLogger) *PruneCache {
	tc := &PruneCache{
		rep:           rep,
		ledgerStorage: ledgerStorage,
		states:        &states{diffs: make([]*diffLayer, 0), allKeyMap: make(map[string]struct{})},
		logger:        logger,
	}

	p := NewPrunner(rep, ledgerStorage, accountTrieStorage, storageTrieStorage, tc.states, logger)
	tc.prunner = p
	go p.pruning()
	return tc
}

func (tc *PruneCache) addNewDifflayer(height uint64, ledgerStorage storage.Storage, trieJournals types.TrieJournalBatch, persist bool) {
	l := &diffLayer{
		height:        height,
		ledgerStorage: ledgerStorage,
		cache:         make(map[string]types.Node),
		accountCache:  make(map[string]types.Node),
		storageCache:  make(map[string]types.Node),
	}
	if persist {
		batch := l.ledgerStorage.NewBatch()
		batch.Put(utils.CompositeKey(utils.TrieJournalKey, height), trieJournals.Encode())
		batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MaxHeightStr), utils.MarshalHeight(height))
		if height == 1 {
			batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr), utils.MarshalHeight(height))
		}
		batch.Commit()
	}

	for _, journal := range trieJournals {
		for k := range journal.PruneSet {
			l.cache[k] = nil
			if journal.Type == TypeAccount {
				l.accountCache[k] = nil
			} else {
				l.storageCache[k] = nil
			}
			tc.states.allKeyMap[k] = struct{}{}
		}
		for k, v := range journal.DirtySet {
			l.cache[k] = v
			if journal.Type == TypeAccount {
				l.accountCache[k] = v
			} else {
				l.storageCache[k] = v
			}
			tc.states.allKeyMap[k] = struct{}{}
		}
	}
	tc.states.diffs = append(tc.states.diffs, l)
	tc.states.size.Add(1)
}

func (tc *PruneCache) Update(height uint64, trieJournals types.TrieJournalBatch) {
	tc.states.lock.Lock()
	defer tc.states.lock.Unlock()

	tc.logger.Debugf("[PruneCache-Update] update trie cache at height: %v, journal=%v", height, trieJournals)

	tc.addNewDifflayer(height, tc.ledgerStorage, trieJournals, true)

}

func (tc *PruneCache) Get(version uint64, key []byte) (res types.Node, ok bool) {
	tc.states.lock.RLock()
	defer tc.states.lock.RUnlock()

	if len(tc.states.diffs) == 0 {
		return nil, false
	}

	if _, ok = tc.states.allKeyMap[string(key)]; !ok {
		return nil, false
	}

	// TODO: use bloom filter to accelerate querying
	for i := len(tc.states.diffs) - 1; i >= 0; i-- {
		if tc.states.diffs[i].height > version {
			continue
		}
		if res, ok = tc.states.diffs[i].GetFromTrieCache(key); ok {
			break
		}
	}

	//tc.logger.Infof("[PruneCache-Get] get from PruneCache: version=%v,key=%v,ok=%v", version, key, ok)

	return res, ok
}

func (tc *PruneCache) Rollback(height uint64) error {
	tc.states.lock.Lock()
	defer tc.states.lock.Unlock()

	minHeight, maxHeight := tc.GetRange()

	tc.logger.Infof("[PruneCache-Rollback] minHeight=%v,maxHeight=%v, targetHeight=%v", minHeight, maxHeight, height)

	// empty cache, no-op
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

	tc.states.diffs = make([]*diffLayer, 0)
	tc.states.allKeyMap = make(map[string]struct{}, 0)
	tc.states.size.Store(0)

	batch := tc.ledgerStorage.NewBatch()
	for i := minHeight; i <= height; i++ {
		trieJournal := tc.GetTrieJournal(i)
		tc.logger.Debugf("[PruneCache-Rollback] apply trie journal of height=%v, trieJournal=%v", i, trieJournal)
		if trieJournal == nil {
			break
		}
		tc.addNewDifflayer(i, tc.ledgerStorage, trieJournal, false)
	}
	batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MaxHeightStr), utils.MarshalHeight(height))

	for i := height + 1; i <= maxHeight; i++ {
		batch.Delete(utils.CompositeKey(utils.TrieJournalKey, i))
	}

	batch.Commit()
	tc.states.rebuildAllKeyMap()

	return nil
}

func (tc *PruneCache) GetRange() (uint64, uint64) {
	minHeight := uint64(0)
	maxHeight := uint64(0)

	data := tc.ledgerStorage.Get(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr))
	if data != nil {
		minHeight = utils.UnmarshalHeight(data)
	}

	data = tc.ledgerStorage.Get(utils.CompositeKey(utils.TrieJournalKey, utils.MaxHeightStr))
	if data != nil {
		maxHeight = utils.UnmarshalHeight(data)
	}

	return minHeight, maxHeight
}

func (tc *PruneCache) GetTrieJournal(height uint64) types.TrieJournalBatch {
	data := tc.ledgerStorage.Get(utils.CompositeKey(utils.TrieJournalKey, height))
	if data == nil {
		return nil
	}

	res, err := types.DecodeTrieJournalBatch(data)
	if err != nil {
		panic(err)
	}

	return res
}
