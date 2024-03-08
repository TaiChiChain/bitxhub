package prune

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
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
	maxFlushBlockNum               = 240
	checkFlushTimeInterval         = 5 * time.Second
	maxFlushTimeInterval           = 2 * time.Minute
	maxFlushBatchSizeThreshold     = 12 * 1024 * 1024 // 12MB
)

func NewPrunner(rep *repo.Repo, ledgerStorage storage.Storage, states *states, logger logrus.FieldLogger) *prunner {
	return &prunner{
		rep:           rep,
		ledgerStorage: ledgerStorage,
		states:        states,
		logger:        logger,
		lastPruneTime: time.Now(),
	}
}

// todo configure different pruning strategies
func (p *prunner) pruning() {
	p.logger.Infof("[Prune] start prunner")
	reserve := defaultMinimumReservedBlockNum
	if p.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum > reserve {
		reserve = p.rep.Config.Ledger.StateLedgerReservedHistoryBlockNum
	}

	var (
		ticker               = time.NewTicker(checkFlushTimeInterval)
		pendingBatch         = p.ledgerStorage.NewBatch()
		from                 uint64 // block from
		to                   uint64 // block to
		pendingFlushBlockNum uint32
	)

	flushFunc := func() {
		if pendingFlushBlockNum <= 0 || pendingBatch.Size() <= 0 {
			return
		}
		pendingBatch.Put(utils.CompositeKey(utils.TrieJournalKey, utils.MinHeightStr), utils.MarshalHeight(to+1))
		pendingBatch.Commit()

		//reset states diff
		{
			p.states.lock.Lock()
			defer p.states.lock.Unlock()
			p.states.diffs = p.states.diffs[pendingFlushBlockNum:]
			p.states.size.Add(int32(-pendingFlushBlockNum))
			p.logger.Infof("[Prune] prune state from block %v to block %v", from, to)
		}

		// reset tmp var
		pendingBatch = p.ledgerStorage.NewBatch()
		from = 0
		to = 0
		pendingFlushBlockNum = 0
		p.lastPruneTime = time.Now()
	}

	for {
		select {
		case <-ticker.C:
			if int(p.states.size.Load()) <= reserve {
				break
			}
			pendingStales := p.states.diffs[pendingFlushBlockNum : len(p.states.diffs)-reserve]
			if len(pendingStales) > 0 {
				if from == 0 {
					from = pendingStales[0].height
				}
				to = pendingStales[len(pendingStales)-1].height

				pruneSet := make(map[string]struct{})
				writeSet := make(map[string]struct{})
				// merge prune set and write set, reduce duplicated entries
				for _, diff := range pendingStales {
					for _, journal := range diff.trieJournals {
						for k := range journal.PruneSet {
							pruneSet[k] = struct{}{}
						}
						for k := range journal.DirtySet {
							writeSet[k] = struct{}{}
						}
					}

				}
				for _, diff := range pendingStales {
					for _, journal := range diff.trieJournals {
						for k := range journal.PruneSet {
							if _, ok := writeSet[k]; !ok {
								pendingBatch.Delete([]byte(k))
							}
						}
						for k, v := range journal.DirtySet {
							if _, ok := pruneSet[k]; !ok {
								pendingBatch.Put([]byte(k), v)
							}
						}
					}
					pendingBatch.Delete(utils.CompositeKey(utils.TrieJournalKey, diff.height))
					pendingFlushBlockNum++
				}
			}

			if time.Since(p.lastPruneTime) >= maxFlushTimeInterval || pendingFlushBlockNum >= maxFlushBlockNum || pendingBatch.Size() > maxFlushBatchSizeThreshold {
				flushFunc()
			}
		}
	}
}
