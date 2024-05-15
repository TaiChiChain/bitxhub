package status

import (
	"sync/atomic"
)

type StatusType uint32

type StatusMgr struct {
	atomicStatus uint32
}

func NewStatusMgr() *StatusMgr {
	return &StatusMgr{
		atomicStatus: 0,
	}
}

// turn on a status
func (st *StatusMgr) On(statusPos ...StatusType) {
	for _, pos := range statusPos {
		st.atomicSetBit(pos)
	}
}

func (st *StatusMgr) atomicSetBit(position StatusType) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&st.atomicStatus)
		if atomic.CompareAndSwapUint32(&st.atomicStatus, oldStatus, oldStatus|(1<<position)) {
			break
		}
	}
}

// turn off a status
func (st *StatusMgr) Off(status ...StatusType) {
	for _, pos := range status {
		st.atomicClearBit(pos)
	}
}

func (st *StatusMgr) atomicClearBit(position StatusType) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&st.atomicStatus)
		if atomic.CompareAndSwapUint32(&st.atomicStatus, oldStatus, oldStatus&^(1<<position)) {
			break
		}
	}
}

// In returns the atomic status of specified position.
func (st *StatusMgr) In(pos StatusType) bool {
	return st.atomicHasBit(pos)
}

func (st *StatusMgr) atomicHasBit(position StatusType) bool {
	val := atomic.LoadUint32(&st.atomicStatus) & (1 << position)
	return val > 0
}

func (st *StatusMgr) InOne(poss ...StatusType) bool {
	var rs = false
	for _, pos := range poss {
		rs = rs || st.In(pos)
	}
	return rs
}

func (st *StatusMgr) Reset() {
	atomic.StoreUint32(&st.atomicStatus, 0)
}
