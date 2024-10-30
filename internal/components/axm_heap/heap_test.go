package axm_heap

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	q := &NumberHeap{}
	heap.Init(q)

	q.PushItem(1)
	q.PushItem(19)
	q.PushItem(12)
	fmt.Println(q.PopItem())
	fmt.Println(q.PopItem())
	fmt.Println(q.PopItem())
	fmt.Println(q.PopItem())
	fmt.Println(q.PopItem())
	q.PushItem(16)
	fmt.Println(q.PopItem())
}
