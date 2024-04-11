package prune

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// PruneCache enables trie node caches, so that every trie read op will happen in trie cache first,
// which avoids frequent reading from disk.
// By using PruneCache, we also implement trie pruning schema, which reduces state storage size that a full node must hold.
type PruneCache struct {
	rep           *repo.Repo
	ledgerStorage storage.Storage
	states        *states

	prunner *prunner

	logger logrus.FieldLogger
}

type states struct {
	diffs     []*diff
	lock      sync.RWMutex
	allKeyMap map[string]struct{}
}

type diff struct {
	height uint64

	accountDiff map[string]types.Node
	storageDiff map[string]types.Node

	ledgerStorage storage.Storage
}

const (
	TypeAccount = 1
	TypeStorage = 2
)

var (
	ErrorRollbackToHigherNumber = errors.New("rollback PruneCache to higher blockchain height")
	ErrorRollbackTooMuch        = errors.New("rollback PruneCache too much block")
)

func (s *states) rebuildAllKeyMap() {
	s.allKeyMap = make(map[string]struct{})
	if len(s.diffs) > 0 {
		for _, diff := range s.diffs {
			for k := range diff.accountDiff {
				s.allKeyMap[k] = struct{}{}
			}
			for k := range diff.storageDiff {
				s.allKeyMap[k] = struct{}{}
			}
		}
	}
}

func NewPruneCache(rep *repo.Repo, ledgerStorage storage.Storage, accountTrieCache *storagemgr.CacheWrapper, storageTrieCache *storagemgr.CacheWrapper, logger logrus.FieldLogger) *PruneCache {
	tc := &PruneCache{
		rep:           rep,
		ledgerStorage: ledgerStorage,
		states:        &states{diffs: make([]*diff, 0), allKeyMap: make(map[string]struct{})},
		logger:        logger,
	}

	p := NewPrunner(rep, ledgerStorage, accountTrieCache, storageTrieCache, tc.states, logger)
	tc.prunner = p
	if rep.Config.Ledger.EnablePrune {
		go p.pruning()
	}
	return tc
}

func (tc *PruneCache) addNewDiff(batch storage.Batch, height uint64, ledgerStorage storage.Storage, stateDelta *types.StateDelta, persist bool) {
	l := &diff{
		height:        height,
		ledgerStorage: ledgerStorage,
		accountDiff:   make(map[string]types.Node),
		storageDiff:   make(map[string]types.Node),
	}
	if persist {

		err := tc.prunner.blockJournal.Append(height, stateDelta.Encode())
		if err != nil {
			panic(err)
		}
	}

	for _, journal := range stateDelta.Journal {
		batch.Put(journal.RootHash[:], journal.RootNodeKey.Encode())
		for k := range journal.PruneSet {
			if journal.Type == TypeAccount {
				l.accountDiff[k] = nil
			} else {
				l.storageDiff[k] = nil
			}
			tc.states.allKeyMap[k] = struct{}{}
		}
		for k, v := range journal.DirtySet {
			if journal.Type == TypeAccount {
				l.accountDiff[k] = v
			} else {
				l.storageDiff[k] = v
			}
			tc.states.allKeyMap[k] = struct{}{}
		}
	}
	tc.states.diffs = append(tc.states.diffs, l)
}

func (tc *PruneCache) Update(batch storage.Batch, height uint64, trieJournals *types.StateDelta) {
	tc.states.lock.Lock()
	defer tc.states.lock.Unlock()

	tc.logger.Debugf("[PruneCache-Update] update trie cache at height: %v", height)

	tc.addNewDiff(batch, height, tc.ledgerStorage, trieJournals, true)
}

func (tc *PruneCache) Get(version uint64, key []byte) (types.Node, bool) {
	tc.states.lock.RLock()
	defer tc.states.lock.RUnlock()

	if len(tc.states.diffs) == 0 {
		return nil, false
	}

	k := string(key)

	if _, ok := tc.states.allKeyMap[k]; !ok {
		return nil, false
	}

	for i := len(tc.states.diffs) - 1; i >= 0; i-- {
		if tc.states.diffs[i].height > version {
			continue
		}
		// the origin trie node may be recycled later, so we must deep-copy it here.
		if v, ok := tc.states.diffs[i].accountDiff[k]; ok {
			if v == nil {
				return v, ok
			}
			return v.Copy(), true
		}
		if v, ok := tc.states.diffs[i].storageDiff[k]; ok {
			if v == nil {
				return v, ok
			}
			return v.Copy(), true
		}
	}

	return nil, false
}

// Rollback rebuilds pruneCache from pruneJournal at target height.
func (tc *PruneCache) Rollback(height uint64) error {
	tc.states.lock.Lock()
	defer tc.states.lock.Unlock()

	minHeight, maxHeight := tc.prunner.blockJournal.GetJournalRange()

	tc.logger.Infof("[PruneCache-Rollback] minHeight=%v, maxHeight=%v, targetHeight=%v", minHeight, maxHeight, height)

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

	tc.states.diffs = make([]*diff, 0)
	tc.states.allKeyMap = make(map[string]struct{}, 0)

	batch := tc.ledgerStorage.NewBatch()
	for i := minHeight; i <= height; i++ {
		trieJournal, err := tc.GetPruneJournal(i)
		if err != nil {
			return err
		}
		tc.logger.Debugf("[PruneCache-Rollback] apply trie journal of height=%v, trieJournal=%v", i, trieJournal)
		if trieJournal == nil {
			tc.logger.Warnf("[PruneCache-Rollback] trie journal is empty at height: %v", i)
			continue
		}
		tc.addNewDiff(batch, i, tc.ledgerStorage, trieJournal, false)
	}

	err := tc.prunner.blockJournal.Truncate(height)
	if err != nil {
		return err
	}

	batch.Commit()
	tc.states.rebuildAllKeyMap()

	return nil
}

func (tc *PruneCache) GetRange() (uint64, uint64) {
	return tc.prunner.blockJournal.GetJournalRange()
}

func (tc *PruneCache) GetPruneJournal(height uint64) (*types.StateDelta, error) {
	data, err := tc.prunner.blockJournal.Retrieve(height)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	res, err := types.DecodeStateDelta(data)
	if err != nil {
		panic(err)
	}

	return res, nil
}

// for debug
func (dl *diff) String() string {
	res := strings.Builder{}
	res.WriteString("Version[")
	res.WriteString(strconv.Itoa(int(dl.height)))
	res.WriteString("], \nAccountDiff[\n")
	res.WriteString(fmt.Sprintf("journal=%v\n", dl.accountDiff))
	res.WriteString("], \nStorageDiff[\n")
	res.WriteString(fmt.Sprintf("journal=%v\n", dl.storageDiff))
	res.WriteString("]")
	return res.String()
}
