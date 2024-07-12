package prune

import (
	"github.com/axiomesh/axiom-kit/jmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type prunner struct {
	rep    *repo.Repo
	states *states

	ledgerStorageBackend kv.Storage
	//accountTrieCache     jmt.TrieCache
	//storageTrieCache     jmt.TrieCache
	trieCache jmt.TrieCache

	logger logrus.FieldLogger

	lastPruneTime time.Time
}

var (
	defaultMinimumReservedBlockNum = 2
	checkFlushTimeInterval         = 10 * time.Second
	maxFlushBlockNum               = 10
	maxFlushTimeInterval           = 1 * time.Minute
	maxFlushBatchSizeThreshold     = 32 * 1024 * 1024 // 32MB
)

func NewPrunner(rep *repo.Repo, ledgerStorage kv.Storage, trieCache jmt.TrieCache, states *states, logger logrus.FieldLogger) *prunner {
	return &prunner{
		rep:                  rep,
		ledgerStorageBackend: ledgerStorage,
		//accountTrieCache:     accountTrieCache,
		//storageTrieCache:     storageTrieCache,
		trieCache:     trieCache,
		states:        states,
		logger:        logger,
		lastPruneTime: time.Now(),
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
		accountTriePruneSet, storageTriePruneSet = make(map[string]*jmt.NodeData), make(map[string]*jmt.NodeData)
		accountTrieWriteSet, storageTrieWriteSet = make(map[string]*jmt.NodeData), make(map[string]*jmt.NodeData)
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
						accountTriePruneSet[k] = v
						pendingFlushSize += len(k)
					} else {
						accountTrieWriteSet[k] = v
						pendingFlushSize += len(k)
					}
				}
				// handle storage trie cache
				for k, v := range diff.storageDiff {
					if v == nil {
						storageTriePruneSet[k] = v
						pendingFlushSize += len(k)
					} else {
						storageTrieWriteSet[k] = v
						pendingFlushSize += len(k)
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

		insertNodes, delNodes := make([]*jmt.NodeData, 0), make([]*jmt.NodeData, 0)
		// update account trie cache
		for k, v := range accountTrieWriteSet {
			if _, has := accountTriePruneSet[k]; !has {
				pendingBatch.Put([]byte(k), v.Node.Encode())
				insertNodes = append(insertNodes, v)
			}
		}
		for k, v := range accountTriePruneSet {
			if _, has := accountTrieWriteSet[k]; !has {
				pendingBatch.Delete([]byte(k))
				delNodes = append(delNodes, v)
			}
		}

		// update storage trie cache
		for k, v := range storageTrieWriteSet {
			if _, has := storageTriePruneSet[k]; !has {
				pendingBatch.Put([]byte(k), v.Node.Encode())
				insertNodes = append(insertNodes, v)
			}
		}
		for k, v := range storageTriePruneSet {
			if _, has := storageTrieWriteSet[k]; !has {
				pendingBatch.Delete([]byte(k))
				delNodes = append(delNodes, v)
			}
		}
		p.trieCache.(*jmt.JMTCache).BatchDelete(delNodes)
		p.trieCache.(*jmt.JMTCache).BatchInsert(insertNodes)

		pendingBatch.Put(utils.CompositeKey(utils.PruneJournalKey, utils.MinHeightStr), utils.MarshalUint64(to+1))

		current := time.Now()
		pendingBatch.Commit()
		p.logger.Infof("[Prune] prune state from block %v to block %v, total size (bytes) = %v, time = %v", from, to, pendingFlushSize, time.Since(current))

		// reset states diff
		p.states.lock.Lock()
		stales := p.states.diffs[:pendingFlushBlockNum]
		for _, d := range stales {
			for _, node := range d.accountDiff {
				if node == nil {
					continue
				}
				types.RecycleTrieNode(node.Node)
			}
			for _, node := range d.storageDiff {
				if node == nil {
					continue
				}
				types.RecycleTrieNode(node.Node)
			}
		}
		p.states.diffs = p.states.diffs[pendingFlushBlockNum:]
		p.states.rebuildAllKeyMap()
		p.states.lock.Unlock()

		pendingBatch.Reset()
		from, to = 0, 0
		p.lastPruneTime = time.Now()
		accountTriePruneSet, storageTriePruneSet = make(map[string]*jmt.NodeData), make(map[string]*jmt.NodeData)
		accountTrieWriteSet, storageTrieWriteSet = make(map[string]*jmt.NodeData), make(map[string]*jmt.NodeData)
		pendingFlushBlockNum, pendingFlushSize = 0, 0
	}
}
