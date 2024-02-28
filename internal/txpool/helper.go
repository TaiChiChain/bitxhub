package txpool

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (p *txPoolImpl[T, Constraint]) setFull() {
	p.statusMgr.On(PoolFull)
}

func (p *txPoolImpl[T, Constraint]) setNotFull() {
	p.statusMgr.Off(PoolFull)
}

func (p *txPoolImpl[T, Constraint]) setReady() {
	p.statusMgr.On(ReadyGenerateBatch)
}

func (p *txPoolImpl[T, Constraint]) setNotReady() {
	p.statusMgr.Off(ReadyGenerateBatch)
}

func (p *txPoolImpl[T, Constraint]) setHasPendingRequest() {
	p.statusMgr.On(HasPendingRequest)
}

func (p *txPoolImpl[T, Constraint]) setNoPendingRequest() {
	p.statusMgr.Off(HasPendingRequest)
}

func traceRejectTx(reason string) {
	rejectTxNum.With(prometheus.Labels{"reason": reason}).Inc()
	rejectTxNum.With(prometheus.Labels{"reason": "all"}).Inc()
}

func traceRemovedTx(reason string, count int) {
	removeTxNum.With(prometheus.Labels{"reason": reason}).Add(float64(count))
	removeTxNum.With(prometheus.Labels{"reason": "all"}).Add(float64(count))
}

func traceProcessEvent(event string, duration time.Duration) {
	processEventDuration.With(prometheus.Labels{"event": event}).Observe(duration.Seconds())
	processEventDuration.With(prometheus.Labels{"event": "all"}).Observe(duration.Seconds())
}

func tracePersistRecords(duration time.Duration) {
	insertRecordDuration.Observe(duration.Seconds())
}
