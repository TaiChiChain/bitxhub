package prune

import (
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type prunner struct {
	rep    *repo.Repo
	states *states

	ledgerStorageBackend storage.Storage
	accountTrieStorage   *storagemgr.CachedStorage
	storageTrieStorage   *storagemgr.CachedStorage

	logger logrus.FieldLogger

	lastPruneTime time.Time
}

const (
	defaultMinimumReservedBlockNum = 2
	maxFlushBlockNum               = 10
	checkFlushTimeInterval         = 5 * time.Second
	maxFlushTimeInterval           = 5 * time.Minute
	maxFlushBatchSizeThreshold     = 12 * 1024 * 1024 // 12MB
)

func NewPrunner(rep *repo.Repo, ledgerStorage storage.Storage, accountTrieStorage *storagemgr.CachedStorage, storageTrieStorage *storagemgr.CachedStorage, states *states, logger logrus.FieldLogger) *prunner {
	return &prunner{
		rep:                  rep,
		ledgerStorageBackend: ledgerStorage,
		accountTrieStorage:   accountTrieStorage,
		storageTrieStorage:   storageTrieStorage,
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
		ticker               = time.NewTicker(checkFlushTimeInterval)
		pendingBatch         = p.ledgerStorageBackend.NewBatch()
		from                 uint64 // block from
		to                   uint64 // block to
		pruneSet             = make(map[string]struct{})
		writeSet             = make(map[string]struct{})
		pendingFlushBlockNum int
	)

	for {
		select {
		case <-ticker.C:
			p.states.lock.RLock()
			if int(p.states.size.Load()) <= reserve {
				p.states.lock.RUnlock()
				break
			}

			if pendingFlushBlockNum > len(p.states.diffs)-reserve {
				p.states.lock.RUnlock()
				break
			}

			pendingStales := p.states.diffs[pendingFlushBlockNum : len(p.states.diffs)-reserve]
			if len(pendingStales) > 0 {
				if from == 0 {
					from = pendingStales[0].height
				}
				to = pendingStales[len(pendingStales)-1].height

				// merge prune set and write set, reduce duplicated entries
				for _, diff := range pendingStales {
					for k, v := range diff.cache {
						if v == nil {
							pruneSet[k] = struct{}{}
						} else {
							writeSet[k] = struct{}{}
						}
					}
				}
				// todo confirm concurrent rw safety here
				for _, diff := range pendingStales {
					for k, v := range diff.cache {
						if v == nil {
							if _, ok := writeSet[k]; !ok {
								if _, has := diff.accountCache[k]; has {
									p.accountTrieStorage.PutCache([]byte(k), nil)
								} else if _, has = diff.storageCache[k]; has {
									p.storageTrieStorage.PutCache([]byte(k), nil)
								}
								pendingBatch.Delete([]byte(k))
							}
						} else {
							if _, ok := pruneSet[k]; !ok {
								if _, has := diff.accountCache[k]; has {
									p.accountTrieStorage.PutCache([]byte(k), v.EncodePb())
								} else if _, has = diff.storageCache[k]; has {
									p.storageTrieStorage.PutCache([]byte(k), v.EncodePb())
								}
								pendingBatch.Put([]byte(k), v.EncodePb())
							}
						}
					}

					pendingBatch.Delete(utils.CompositeKey(utils.TrieJournalKey, diff.height))
				}
				pendingFlushBlockNum += len(pendingStales)
			}
			p.states.lock.RUnlock()

			if time.Since(p.lastPruneTime) >= maxFlushTimeInterval || pendingFlushBlockNum >= maxFlushBlockNum || pendingBatch.Size() >= maxFlushBatchSizeThreshold {
				pendingBatch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr), utils.MarshalHeight(to+1))
				pendingBatch.Commit()

				//reset states diff
				p.states.lock.Lock()
				p.states.diffs = p.states.diffs[pendingFlushBlockNum:]
				p.states.size.Add(int32(-pendingFlushBlockNum))
				p.states.rebuildAllKeyMap()
				p.states.lock.Unlock()

				pendingBatch.Reset()
				from, to = 0, 0
				p.lastPruneTime = time.Now()
				pruneSet = make(map[string]struct{})
				writeSet = make(map[string]struct{})
				pendingFlushBlockNum = 0
				p.logger.Infof("[Prune] prune state from block %v to block %v", from, to)
			}
		}
	}
}
