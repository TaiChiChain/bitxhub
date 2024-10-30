package axm_heap

import (
	"container/heap"
	"math"
)

// NumberHeap a heap that sort by number (ascending)
type NumberHeap []uint64

// PeekItem returns the top element of the NumberHeap without removing it.
func (n *NumberHeap) PeekItem() uint64 {
	if n.Len() == 0 {
		return math.MaxUint64
	}

	return (*n)[0]
}

// PushItem adds a new element to the NumberHeap.
func (n *NumberHeap) PushItem(x uint64) {
	heap.Push(n, x)
}

// PopItem removes and returns the top element from the NumberHeap.
func (n *NumberHeap) PopItem() uint64 {
	if n.Len() == 0 {
		return math.MaxUint64
	}
	return heap.Pop(n).(uint64)
}

/* Queue methods required by the heap interface */

func (n *NumberHeap) Len() int {
	return len(*n)
}

func (n *NumberHeap) Swap(i, j int) {
	(*n)[i], (*n)[j] = (*n)[j], (*n)[i]
}

func (n *NumberHeap) Less(i, j int) bool {
	return (*n)[i] < (*n)[j]
}

func (n *NumberHeap) Push(x any) {
	nonce, ok := x.(uint64)
	if !ok {
		return
	}

	*n = append(*n, nonce)
}

func (n *NumberHeap) Pop() any {
	old := *n
	num := len(old)
	item := old[num-1]
	old[num-1] = 0 // avoid memory leak
	*n = old[0 : num-1]

	return item
}
