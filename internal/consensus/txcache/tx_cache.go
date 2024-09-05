package txcache

import (
	"sync"
	"time"

	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
)

const (
	DefaultTxCacheSize = 10000
	DefaultTxSetTick   = 100 * time.Millisecond
	DefaultTxSetSize   = 10
)

type TxCache interface {
	Start()
	Stop()
	PostTx(tx *types.Transaction)
	CommitTxSet() <-chan []*types.Transaction
}

type TxCacheImpl struct {
	txSetC  chan []*types.Transaction
	recvTxC chan *types.Transaction
	txRespC chan *consensustypes.TxWithResp
	closeC  chan bool
	wg      sync.WaitGroup

	metrics    *txCacheMetrics
	txSet      []*types.Transaction
	logger     logrus.FieldLogger
	timerC     chan bool
	stopTimerC chan bool
	txSetTick  time.Duration
	TxSetSize  uint64
}

func NewTxCache(txSliceTimeout time.Duration, txSetSize uint64, logger logrus.FieldLogger) *TxCacheImpl {
	txCache := &TxCacheImpl{}
	txCache.recvTxC = make(chan *types.Transaction, DefaultTxCacheSize)
	txCache.txSetC = make(chan []*types.Transaction)
	txCache.txRespC = make(chan *consensustypes.TxWithResp)
	txCache.timerC = make(chan bool)
	txCache.stopTimerC = make(chan bool)
	txCache.txSet = make([]*types.Transaction, 0)
	txCache.logger = logger
	if txSliceTimeout == 0 {
		txCache.txSetTick = DefaultTxSetTick
	} else {
		txCache.txSetTick = txSliceTimeout
	}
	if txSetSize == 0 {
		txCache.TxSetSize = DefaultTxSetSize
	} else {
		txCache.TxSetSize = txSetSize
	}

	txCache.logger.Infof("[TxCacheImpl Init] set size = %d", txCache.TxSetSize)
	txCache.logger.Infof("[TxCacheImpl Init] set timeout = %v", txCache.txSetTick)
	txCache.logger.Infof("[TxSet Init] channel size = %d", DefaultTxCacheSize)
	return txCache
}

func (tc *TxCacheImpl) Start() {
	// todo: refactor metrics
	tc.metrics = newTxCacheMetrics()
	tc.closeC = make(chan bool)

	tc.wg.Add(1)
	go tc.listenEvent()
}

func (tc *TxCacheImpl) CommitTxSet() <-chan []*types.Transaction {
	return tc.txSetC
}

func (tc *TxCacheImpl) PostTx(tx *types.Transaction) {
	tc.metrics.pendingTxs.Add(float64(1))
	tc.recvTxC <- tx
}

func (tc *TxCacheImpl) listenEvent() {
	defer tc.wg.Done()
	for {
		select {
		case <-tc.closeC:
			tc.logger.Info("Transaction cache stopped!")
			return

		case tx := <-tc.recvTxC:
			tc.metrics.incomingTxsCounter.Add(float64(1))
			tc.appendTx(tx)

		case <-tc.timerC:
			tc.stopTxSetTimer()
			tc.postTxSet()
		}
	}
}

func (tc *TxCacheImpl) appendTx(tx *types.Transaction) {
	if tx == nil {
		tc.logger.Errorf("Transaction is nil")
		return
	}
	if len(tc.txSet) == 0 {
		tc.startTxSetTimer()
	}
	tc.txSet = append(tc.txSet, tx)
	if uint64(len(tc.txSet)) >= tc.TxSetSize {
		tc.stopTxSetTimer()
		tc.postTxSet()
	}
}

func (tc *TxCacheImpl) Stop() {
	common.DrainChannel[*types.Transaction](tc.recvTxC)
	if tc.closeC != nil {
		close(tc.closeC)
	}
	tc.closeC = nil
	tc.wg.Wait()
	tc.stopTxSetTimer()
	tc.txSet = nil
	tc.logger.Info("Transaction cache stopped!")
}

func (tc *TxCacheImpl) postTxSet() {
	dst := make([]*types.Transaction, len(tc.txSet))
	copy(dst, tc.txSet)
	tc.txSet = make([]*types.Transaction, 0)
	tc.txSetC <- dst
	var (
		start                                      = time.Now().UnixNano()
		maxLatency, minLatency, totalLatency int64 = 0, start, 0
	)

	for _, e := range dst {
		latency := start - e.RbftGetTimeStamp()
		if latency > maxLatency {
			maxLatency = latency
		}
		if latency < minLatency {
			minLatency = latency
		}
		totalLatency += latency
	}

	tc.metrics.txAPILatency.WithLabelValues("max").Observe(time.Duration(maxLatency).Seconds())
	tc.metrics.txAPILatency.WithLabelValues("min").Observe(time.Duration(minLatency).Seconds())
	tc.metrics.txAPILatency.WithLabelValues("avg").Observe(time.Duration(totalLatency).Seconds() / float64(len(dst)))

	tc.metrics.pendingTxs.Add(float64(-len(dst)))
}

func (tc *TxCacheImpl) IsFull() bool {
	return len(tc.recvTxC) == DefaultTxCacheSize
}

func (tc *TxCacheImpl) startTxSetTimer() {
	go func() {
		timer := time.NewTimer(tc.txSetTick)
		select {
		case <-timer.C:
			tc.timerC <- true
		case <-tc.stopTimerC:
			return
		}
	}()
}

func (tc *TxCacheImpl) stopTxSetTimer() {
	close(tc.stopTimerC)
	tc.stopTimerC = make(chan bool)
}
