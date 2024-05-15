package archive

import (
	"container/heap"
	"encoding/hex"
	"math"
	"sync"

	"github.com/axiomesh/axiom-bft/common/consensus"
	axm_heap "github.com/axiomesh/axiom-ledger/internal/components/heap"
	"github.com/sirupsen/logrus"
)

// All methods assume the (correct) lock is held.
type checkpointStore struct {
	items               map[uint64]*diffCheckpoint
	lastPersistedHeight uint64
	size                uint64
	ready               *axm_heap.NumberHeap
	lock                sync.RWMutex
	logger              logrus.FieldLogger
}

type diffCheckpoint struct {
	values       map[string][]*consensus.SignedCheckpoint
	mostMatching mostMatchingIndex
}

type mostMatchingIndex struct {
	count int
	index string
}

func newDiffCheckpoint() *diffCheckpoint {
	return &diffCheckpoint{
		values: make(map[string][]*consensus.SignedCheckpoint),
	}
}

func (d *diffCheckpoint) add(c *consensus.SignedCheckpoint) {
	h := hex.EncodeToString(c.GetCheckpoint().Hash())
	d.values[h] = append(d.values[h], c)
	if len(d.values[h]) > d.mostMatching.count {
		d.mostMatching = mostMatchingIndex{count: len(d.values[h]), index: h}
	}
}

func (d *diffCheckpoint) length(c *consensus.SignedCheckpoint) int {
	h := hex.EncodeToString(c.GetCheckpoint().Hash())
	return len(d.values[h])
}

func (d *diffCheckpoint) getMostMatchingValue() []*consensus.SignedCheckpoint {
	return d.values[d.mostMatching.index]
}

func newCheckpointQueue(logger logrus.FieldLogger) *checkpointStore {
	q := &checkpointStore{
		items:               make(map[uint64]*diffCheckpoint),
		logger:              logger,
		lastPersistedHeight: 0,
		ready:               new(axm_heap.NumberHeap),
	}
	*q.ready = make([]uint64, 0)
	heap.Init(q.ready)

	return q
}

func (q *checkpointStore) moveWatermarks(height uint64) []string {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.lastPersistedHeight = height
	return q.prune(height)
}

func (q *checkpointStore) prune(height uint64) (pruned []string) {
	q.logger.Debugf("pruning checkpoints at height %d", height)

	// 1. remove all checkpoints that are older than height after state updated
	for {
		h := q.ready.PeekItem()
		if h == math.MaxUint64 || h > height {
			break
		}
		q.ready.PopItem()
		if d, exist := q.remove(h); exist {
			pruned = append(pruned, d)
		}
	}
	// 2. remove current checkpoint after executed
	if len(pruned) == 0 {
		if d, exist := q.remove(height); exist {
			pruned = append(pruned, d)
		}
	}
	return
}

func (q *checkpointStore) insert(checkpoint *consensus.SignedCheckpoint) int {
	q.lock.Lock()
	defer q.lock.Unlock()

	// omit older checkpoints and out of range checkpoints
	if checkpoint.Height() <= q.lastPersistedHeight || checkpoint.Height() >= q.lastPersistedHeight+q.size {
		return 0
	}

	item, ok := q.items[checkpoint.Height()]
	if !ok {
		item = newDiffCheckpoint()
	}
	item.add(checkpoint)
	q.items[checkpoint.Height()] = item
	return item.length(checkpoint)
}

func (q *checkpointStore) insertReady(h uint64) {
	q.ready.PushItem(h)
}

func (q *checkpointStore) getItems(h uint64) []*consensus.SignedCheckpoint {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if _, ok := q.items[h]; !ok {
		return nil
	}
	return q.items[h].getMostMatchingValue()
}

func (q *checkpointStore) remove(h uint64) (string, bool) {
	if _, ok := q.items[h]; !ok {
		return "", false
	}
	batchDigest := q.items[h].getMostMatchingValue()[0].GetCheckpoint().GetExecuteState().GetBatchDigest()
	delete(q.items, h)
	return batchDigest, true
}
