package txpool

import (
	"container/heap"
	"math"

	"github.com/axiomesh/axiom-kit/types"
	axm_heap "github.com/axiomesh/axiom-ledger/internal/components/heap"
)

// A thread-safe wrapper of a nonceHeap.
// All methods assume the (correct) lock is held.
type accountQueue[T any, Constraint types.TXConstraint[T]] struct {
	items         map[uint64]*internalTransaction[T, Constraint]
	minNonceQueue *axm_heap.NumberHeap
	lastNonce     uint64
}

func newAccountQueue[T any, Constraint types.TXConstraint[T]](latestNonce uint64) *accountQueue[T, Constraint] {
	q := &accountQueue[T, Constraint]{
		items:         make(map[uint64]*internalTransaction[T, Constraint]),
		minNonceQueue: new(axm_heap.NumberHeap),
		lastNonce:     latestNonce,
	}
	*q.minNonceQueue = make([]uint64, 0)
	heap.Init(q.minNonceQueue)

	return q
}

func (q *accountQueue[T, Constraint]) replaceTx(tx *internalTransaction[T, Constraint]) (*internalTransaction[T, Constraint], bool) {
	nonce := tx.getNonce()
	old, ok := q.items[nonce]
	if ok && old.getNonce() != nonce {
		return nil, false
	}
	q.items[nonce] = tx

	return old, true
}

func (q *accountQueue[T, Constraint]) updateLastNonce(nonce uint64) {
	q.lastNonce = nonce
}

func (q *accountQueue[T, Constraint]) isValidPush(nonce uint64) bool {
	if nonce < q.minNonceQueue.PeekItem() {
		return true
	}

	if nonce != 0 && nonce != q.lastNonce+1 {
		return false
	}
	return true
}

// prune removes all transactions from the minNonceQueue
// with nonce lower than(or equal) given.
func (q *accountQueue[T, Constraint]) prune(nonce uint64) (
	pruned []*internalTransaction[T, Constraint],
) {
	for {
		tx := q.peek()
		if tx == nil || tx.getNonce() == math.MaxUint64 || tx.getNonce() > nonce {
			break
		}

		pruned = append(pruned, q.pop())
	}

	return
}

// push pushes the given transactions onto the minNonceQueue.
func (q *accountQueue[T, Constraint]) push(tx *internalTransaction[T, Constraint]) bool {
	if _, ok := q.items[tx.getNonce()]; ok {
		q.replaceTx(tx)
		return true
	}
	q.items[tx.getNonce()] = tx
	q.minNonceQueue.PushItem(tx.getNonce())
	if q.lastNonce < tx.getNonce() {
		q.lastNonce = tx.getNonce()
	}
	return false
}

// peek returns the first transaction from the minNonceQueue without removing it.
func (q *accountQueue[T, Constraint]) peek() *internalTransaction[T, Constraint] {
	return q.items[q.minNonceQueue.PeekItem()]
}

// pop removes the first transactions from the minNonceQueue and returns it.
func (q *accountQueue[T, Constraint]) pop() *internalTransaction[T, Constraint] {
	if q.length() == 0 {
		return nil
	}

	nonce := q.minNonceQueue.PopItem()

	tx := q.items[nonce]
	delete(q.items, nonce)

	return tx
}

// removeBehind removes all transactions with nonce bigger than given(including the given).
func (q *accountQueue[T, Constraint]) removeBehind(nonce uint64) ([]*internalTransaction[T, Constraint], bool) {
	removedTxs := make([]*internalTransaction[T, Constraint], 0)
	// Otherwise delete the transaction and fix the heap index
	for i := 0; i < q.minNonceQueue.Len(); {
		currentNonce := (*q.minNonceQueue)[i]
		if currentNonce >= nonce {
			heap.Remove(q.minNonceQueue, i)
			removedTxs = append(removedTxs, q.items[currentNonce])
			delete(q.items, currentNonce)
		} else {
			i++
		}
	}

	if nonce == 0 {
		q.lastNonce = 0
	} else {
		q.lastNonce = nonce - 1
	}
	return removedTxs, len(removedTxs) > 0
}

// length returns the number of transactions in the minNonceQueue.
func (q *accountQueue[T, Constraint]) length() uint64 {
	return uint64(q.minNonceQueue.Len())
}

// TxByPriceAndTime implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPriceAndTime[T any, Constraint types.TXConstraint[T]] []*internalTransaction[T, Constraint]

func (tp *TxByPriceAndTime[T, Constraint]) Len() int { return len(*tp) }

func (tp *TxByPriceAndTime[T, Constraint]) Less(i, j int) bool {
	// If the prices are equal, use the time the transaction was first seen for
	// deterministic sorting
	cmp := (*tp)[i].getGasPrice().Cmp((*tp)[j].getGasPrice())
	if cmp == 0 {
		return (*tp)[i].getRawTimestamp() < (*tp)[j].getRawTimestamp()
	}
	return cmp > 0
}

func (tp *TxByPriceAndTime[T, Constraint]) Swap(i, j int) { (*tp)[i], (*tp)[j] = (*tp)[j], (*tp)[i] }

func (tp *TxByPriceAndTime[T, Constraint]) Push(x any) {
	*tp = append(*tp, x.(*internalTransaction[T, Constraint]))
}

func (tp *TxByPriceAndTime[T, Constraint]) push(tx any) {
	heap.Push(tp, tx)
}

func (tp *TxByPriceAndTime[T, Constraint]) Pop() any {
	old := *tp
	n := len(old)
	x := old[n-1]
	old[n-1] = nil
	*tp = old[0 : n-1]
	return x
}

func (tp *TxByPriceAndTime[T, Constraint]) pop() any {
	return heap.Pop(tp)
}

func (tp *TxByPriceAndTime[T, Constraint]) peek() *internalTransaction[T, Constraint] {
	if tp.Len() == 0 {
		return nil
	}
	return (*tp)[0]
}

func (tp *TxByPriceAndTime[T, Constraint]) remove(tx *internalTransaction[T, Constraint]) bool {
	removed := false
	for i := 0; i < tp.Len(); i++ {
		if (*tp)[i] == tx {
			heap.Remove(tp, i)
			removed = true
			break
		}
	}
	return removed
}
