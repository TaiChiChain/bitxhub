package prune

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type prunner struct {
	rep    *repo.Repo
	states *states

	ledgerStorageBackend storage.Storage
	accountTrieCache     *storagemgr.CacheWrapper
	storageTrieCache     *storagemgr.CacheWrapper

	logger logrus.FieldLogger

	lastPruneTime time.Time
}

var (
	defaultMinimumReservedBlockNum = 2
	checkFlushTimeInterval         = 1 * time.Minute
	maxFlushBlockNum               = 128
	maxFlushTimeInterval           = 10 * time.Minute
	maxFlushBatchSizeThreshold     = 32 * 1024 * 1024 // 32MB
)

func NewPrunner(rep *repo.Repo, ledgerStorage storage.Storage, accountTrieCache *storagemgr.CacheWrapper, storageTrieCache *storagemgr.CacheWrapper, states *states, logger logrus.FieldLogger) *prunner {
	return &prunner{
		rep:                  rep,
		ledgerStorageBackend: ledgerStorage,
		accountTrieCache:     accountTrieCache,
		storageTrieCache:     storageTrieCache,
		states:               states,
		logger:               logger,
		lastPruneTime:        time.Now(),
	}
}

func (p *prunner) pruning() {
	p.logger.Infof("[Prune] start prunner")
	reserve := defaultMinimumReservedBlockNum
	if p.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum > reserve {
		reserve = p.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum
	}

	var (
		ticker                                   = time.NewTicker(checkFlushTimeInterval)
		pendingBatch                             = p.ledgerStorageBackend.NewBatch()
		from, to                                 = uint64(0), uint64(0) // block range
		accountTriePruneSet, storageTriePruneSet = make(map[string]struct{}), make(map[string]struct{})
		accountTrieWriteSet, storageTrieWriteSet = make(map[string][]byte), make(map[string][]byte)
		pendingFlushBlockNum, pendingFlushSize   = 0, 0
	)

	for range ticker.C {
		p.states.lock.RLock()
		if len(p.states.diffs) <= reserve || pendingFlushBlockNum > len(p.states.diffs)-reserve {
			p.states.lock.RUnlock()
			continue
		}

		pendingStales := p.states.diffs[pendingFlushBlockNum : len(p.states.diffs)-reserve]
		if len(pendingStales) > 0 {
			if from == 0 {
				from = pendingStales[0].height
			}
			to = pendingStales[len(pendingStales)-1].height

			// merge prune set and write set, reduce duplicated entries
			for _, diff := range pendingStales {
				// handle account trie cache
				for k, v := range diff.accountDiff {
					if v == nil {
						accountTriePruneSet[k] = struct{}{}
						pendingFlushSize += len(k)
					} else {
						blob := v.Encode()
						accountTrieWriteSet[k] = blob
						pendingFlushSize += len(k) + len(blob)
					}
				}
				// handle storage trie cache
				for k, v := range diff.storageDiff {
					if v == nil {
						storageTriePruneSet[k] = struct{}{}
						pendingFlushSize += len(k)
					} else {
						blob := v.Encode()
						storageTrieWriteSet[k] = blob
						pendingFlushSize += len(k) + len(blob)
					}
				}
				pendingBatch.Delete(utils.CompositeKey(utils.PruneJournalKey, diff.height))
			}
			pendingFlushBlockNum += len(pendingStales)
		}
		p.states.lock.RUnlock()

		if time.Since(p.lastPruneTime) < maxFlushTimeInterval && pendingFlushBlockNum < maxFlushBlockNum &&
			pendingFlushSize < maxFlushBatchSizeThreshold {
			continue
		}

		// The moment we update trie cache, other goroutine may read prune cache at the same time.
		// But we don't need to lock here, because the jmt.getNode logic will always try from prune cache first,
		// and we can ensure that the data we update will occur in prune cache.

		// update account trie cache
		for k, v := range accountTrieWriteSet {
			if _, has := accountTriePruneSet[k]; !has {
				pendingBatch.Put([]byte(k), v)
				p.accountTrieCache.Set([]byte(k), v)
			}
		}
		for k := range accountTriePruneSet {
			if _, has := accountTrieWriteSet[k]; !has {
				pendingBatch.Delete([]byte(k))
				p.accountTrieCache.Del([]byte(k))
			}
		}

		// update storage trie cache
		for k, v := range storageTrieWriteSet {
			if _, has := storageTriePruneSet[k]; !has {
				pendingBatch.Put([]byte(k), v)
				p.storageTrieCache.Set([]byte(k), v)
			}
		}
		for k := range storageTriePruneSet {
			if _, has := storageTrieWriteSet[k]; !has {
				pendingBatch.Delete([]byte(k))
				p.storageTrieCache.Del([]byte(k))
			}
		}

		pendingBatch.Put(utils.CompositeKey(utils.PruneJournalKey, utils.MinHeightStr), utils.MarshalHeight(to+1))
		pendingBatch.Commit()

		//reset states diff
		p.states.lock.Lock()
		//stales := p.states.diffs[:pendingFlushBlockNum]
		//for _, d := range stales {
		//	for _, node := range d.accountDiff {
		//		types.RecycleTrieNode(node)
		//	}
		//	for _, node := range d.storageDiff {
		//		types.RecycleTrieNode(node)
		//	}
		//}
		p.states.diffs = p.states.diffs[pendingFlushBlockNum:]
		p.states.rebuildAllKeyMap()
		p.states.lock.Unlock()
		p.logger.Infof("[Prune] prune state from block %v to block %v", from, to)

		pendingBatch.Reset()
		from, to = 0, 0
		p.lastPruneTime = time.Now()
		accountTriePruneSet, storageTriePruneSet = make(map[string]struct{}), make(map[string]struct{})
		accountTrieWriteSet, storageTrieWriteSet = make(map[string][]byte), make(map[string][]byte)
		pendingFlushBlockNum = 0
	}

}
