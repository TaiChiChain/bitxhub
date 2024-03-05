package prune

import (
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
)

type prunner struct {
	rep           *repo.Repo
	ledgerStorage storage.Storage
	states        *states

	logger logrus.FieldLogger

	lastPruneTime time.Time
}

const (
	defaultMinimumReservedBlockNum = 2
	defaultFlushBlockNum           = 16
	defaultFlushTimeInterval       = 1 * time.Second
	defaultFlushBatchSize          = 128 * 1024 * 1024 // 128MB
)

func NewPrunner(rep *repo.Repo, ledgerStorage storage.Storage, states *states, logger logrus.FieldLogger) *prunner {
	return &prunner{
		rep:           rep,
		ledgerStorage: ledgerStorage,
		states:        states,
		logger:        logger,
	}
}

// todo configure different pruning strategies
func (p *prunner) pruning() {
	p.logger.Infof("[Prune] start prunner")
	reserve := defaultMinimumReservedBlockNum
	if p.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum > reserve {
		reserve = p.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum
	}
	ticker := time.NewTicker(defaultFlushTimeInterval)

	for {
		select {
		case <-ticker.C:
			if int(p.states.size.Load()) < defaultFlushBlockNum+reserve {
				break
			}
			stales := p.states.diffs[:defaultFlushBlockNum]
			from, to := stales[0].height, stales[len(stales)-1].height
			batch := p.ledgerStorage.NewBatch()
			pruneSet := make(map[string]struct{})
			writeSet := make(map[string]struct{})
			// merge prune set and write set, reduce duplicated entries
			for _, diff := range stales {
				for _, journal := range diff.trieJournals {
					for k := range journal.PruneSet {
						pruneSet[k] = struct{}{}
					}
					for k := range journal.DirtySet {
						writeSet[k] = struct{}{}
					}
				}
			}
			for _, diff := range stales {
				for _, journal := range diff.trieJournals {
					for k := range journal.PruneSet {
						if _, ok := writeSet[k]; !ok {
							batch.Delete([]byte(k))
						}
					}
					for k, v := range journal.DirtySet {
						if _, ok := pruneSet[k]; !ok {
							batch.Put([]byte(k), v)
						}
					}
				}
				batch.Delete(utils.CompositeKey(utils.TrieJournalKey, diff.height))
			}
			batch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr), utils.MarshalHeight(to+1))
			batch.Commit()

			p.states.lock.Lock()
			p.states.diffs = p.states.diffs[defaultFlushBlockNum:]
			p.states.size.Add(-defaultFlushBlockNum)
			p.states.lock.Unlock()

			p.lastPruneTime = time.Now()
			p.logger.Infof("[Prune] prune state from block %v to block %v", from, to)
		}
	}
}
