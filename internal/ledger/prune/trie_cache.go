package prune

import (
	"errors"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

// TrieCache enables trie node caches, so that every trie read op will happen in trie cache first,
// which avoids frequent reading from disk.
// By using TrieCache, we also implement trie pruning schema, which reduces state storage size that a full node must hold.
type TrieCache struct {
	rep *repo.Repo
	// todo check content of cache, ensure there are only trie nodes
	ledgerStorage storage.Storage
	states        *states

	prunner *prunner

	logger logrus.FieldLogger
}

type states struct {
	size  atomic.Int32
	diffs []*diffLayer
	lock  sync.RWMutex
}

var (
	ErrorRollbackToHigherNumber = errors.New("rollback TrieCache to higher blockchain height")
	ErrorRollbackTooMuch        = errors.New("rollback TrieCache too much block")
)

func NewTrieCache(rep *repo.Repo, ledgerStorage storage.Storage, logger logrus.FieldLogger) *TrieCache {
	tc := &TrieCache{
		rep:           rep,
		ledgerStorage: ledgerStorage,
		states:        &states{diffs: make([]*diffLayer, 0)},
		logger:        logger,
	}

	p := NewPrunner(rep, ledgerStorage, tc.states, logger)
	tc.prunner = p
	go p.pruning()
	return tc
}

func (tc *TrieCache) Update(height uint64, trieJournals types.TrieJournalBatch) {
	tc.states.lock.Lock()
	defer tc.states.lock.Unlock()

	tc.logger.Debugf("[TrieCache-Update] update trie cache at height: %v, journal=%v", height, trieJournals)

	diff := NewDiffLayer(height, tc.ledgerStorage, trieJournals, true)
	tc.states.diffs = append(tc.states.diffs, diff)
	tc.states.size.Add(1)
}

func (tc *TrieCache) Get(version uint64, key []byte) (res []byte, ok bool) {
	tc.states.lock.RLock()
	defer tc.states.lock.RUnlock()

	if len(tc.states.diffs) == 0 {
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

	//tc.logger.Infof("[TrieCache-Get] get from TrieCache: version=%v,key=%v,ok=%v", version, key, ok)

	return res, ok
}

func (tc *TrieCache) Rollback(height uint64) error {
	tc.states.lock.Lock()
	defer tc.states.lock.Unlock()

	minHeight, maxHeight := tc.GetRange()

	tc.logger.Infof("[TrieCache-Rollback] minHeight=%v,maxHeight=%v, targetHeight=%v", minHeight, maxHeight, height)

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
	tc.states.size.Store(0)

	batch := tc.ledgerStorage.NewBatch()
	for i := minHeight; i <= height; i++ {
		trieJournal := tc.GetTrieJournal(i)
		tc.logger.Debugf("[TrieCache-Rollback] apply trie journal of height=%v, trieJournal=%v", i, trieJournal)
		if trieJournal == nil {
			break
		}
		diff := NewDiffLayer(i, tc.ledgerStorage, trieJournal, false)
		tc.states.diffs = append(tc.states.diffs, diff)
		tc.states.size.Add(1)
	}
	batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MaxHeightStr), utils.MarshalHeight(height))

	for i := height + 1; i <= maxHeight; i++ {
		batch.Delete(utils.CompositeKey(utils.TrieJournalKey, i))
	}

	batch.Commit()

	return nil
}

func (tc *TrieCache) GetRange() (uint64, uint64) {
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

func (tc *TrieCache) GetTrieJournal(height uint64) types.TrieJournalBatch {
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
