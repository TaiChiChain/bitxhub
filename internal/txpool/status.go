package txpool

import (
	"sync/atomic"
)

type statusType uint32

const (
	ReadyGenerateBatch statusType = iota
	HasPendingRequest
	PoolFull
	PoolEmpty
)

type PoolStatusMgr struct {
	atomicStatus uint32
}

func newPoolStatusMgr() *PoolStatusMgr {
	return &PoolStatusMgr{
		atomicStatus: 0,
	}
}

// turn on a status
func (st *PoolStatusMgr) On(statusPos ...statusType) {
	for _, pos := range statusPos {
		st.atomicSetBit(pos)
	}
}

func (st *PoolStatusMgr) atomicSetBit(position statusType) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&st.atomicStatus)
		if atomic.CompareAndSwapUint32(&st.atomicStatus, oldStatus, oldStatus|(1<<position)) {
			break
		}
	}
}

// turn off a status
func (st *PoolStatusMgr) Off(status ...statusType) {
	for _, pos := range status {
		st.atomicClearBit(pos)
	}
}

func (st *PoolStatusMgr) atomicClearBit(position statusType) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&st.atomicStatus)
		if atomic.CompareAndSwapUint32(&st.atomicStatus, oldStatus, oldStatus&^(1<<position)) {
			break
		}
	}
}

// In returns the atomic status of specified position.
func (st *PoolStatusMgr) In(pos statusType) bool {
	return st.atomicHasBit(pos)
}

func (st *PoolStatusMgr) atomicHasBit(position statusType) bool {
	val := atomic.LoadUint32(&st.atomicStatus) & (1 << position)
	return val > 0
}

func (st *PoolStatusMgr) InOne(poss ...statusType) bool {
	var rs = false
	for _, pos := range poss {
		rs = rs || st.In(pos)
	}
	return rs
}

func (st *PoolStatusMgr) Reset() {
	atomic.StoreUint32(&st.atomicStatus, 0)
}
