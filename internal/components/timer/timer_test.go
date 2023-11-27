package timer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
)

var (
	eventCh = make(chan any, 1)
	handler = func(name TimeoutEvent) {
		eventCh <- name
	}
)

func resetCh() {
	eventCh = make(chan any, 1)
}

func TestBatchTimer_StartBatchTimer(t *testing.T) {
	logger := log.NewWithModule("timer")
	tm := NewTimerManager(logger)
	defer resetCh()

	var (
		batchEnd     time.Duration
		noTxBatchEnd time.Duration
	)
	batchTimeout := 500 * time.Millisecond
	noTxBatchTimeout := 1 * time.Second
	err := tm.CreateTimer(Batch, 0, handler)
	require.NotNil(t, err)

	err = tm.CreateTimer(Batch, batchTimeout, handler)
	require.Nil(t, err)

	start := time.Now()

	batchEventCh := make(chan struct{})
	noTxBatchEventCh := make(chan struct{})
	defer func() {
		close(eventCh)
		close(batchEventCh)
		close(noTxBatchEventCh)
	}()
	err = tm.StartTimer(Batch)
	require.Nil(t, err)
	err = tm.StartTimer(NoTxBatch)
	require.NotNil(t, err)

	err = tm.CreateTimer(NoTxBatch, noTxBatchTimeout, handler)
	require.Nil(t, err)
	err = tm.StartTimer(NoTxBatch)
	require.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-eventCh:
				switch e {
				case Batch:
					batchEnd = time.Since(start)
					batchEventCh <- struct{}{}
				case NoTxBatch:
					noTxBatchEnd = time.Since(start)
					noTxBatchEventCh <- struct{}{}
				}
			}
		}
	}()
	<-batchEventCh
	<-noTxBatchEventCh
	cancel()
	require.True(t, batchEnd >= batchTimeout)
	require.True(t, noTxBatchEnd >= noTxBatchTimeout)
}

func TestBatchTimer_StopBatchTimer(t *testing.T) {
	defer resetCh()
	logger := log.NewWithModule("timer")
	batchTimer := NewTimerManager(logger)
	batchTimeout := 100 * time.Millisecond
	noTxBatchTimeout := 200 * time.Millisecond
	err := batchTimer.CreateTimer(Batch, batchTimeout, handler)
	require.Nil(t, err)
	err = batchTimer.CreateTimer(NoTxBatch, noTxBatchTimeout, handler)
	require.Nil(t, err)
	ch := make(chan struct{})
	defer func() {
		close(eventCh)
		close(ch)
	}()
	start := time.Now()
	err = batchTimer.StartTimer(Batch)
	require.Nil(t, err)
	err = batchTimer.StartTimer(NoTxBatch)
	require.Nil(t, err)
	go func() {
		select {
		case e := <-eventCh:
			switch e {
			case Batch:
				require.Fail(t, "batch timer should not be triggered")
			case NoTxBatch:
				require.True(t, time.Since(start) >= noTxBatchTimeout)
				ch <- struct{}{}
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)
	batchTimer.StopTimer(Batch)
	require.False(t, batchTimer.IsTimerActive(Batch))
	require.True(t, batchTimer.IsTimerActive(NoTxBatch))
	<-ch
	batchTimer.Stop()
	require.False(t, batchTimer.IsTimerActive(NoTxBatch))
}

func TestBatchTimer_RestartTimer(t *testing.T) {
	defer resetCh()
	logger := log.NewWithModule("timer")
	batchTimer := NewTimerManager(logger)
	batchTimeout := 10 * time.Millisecond
	err := batchTimer.CreateTimer(Batch, batchTimeout, handler)
	require.Nil(t, err)
	ch := make(chan struct{})

	go func() {
		select {
		case e := <-eventCh:
			switch e {
			case Batch:
				ch <- struct{}{}
			}
		}
	}()

	start := time.Now()
	waitTime := 1 * time.Millisecond
	err = batchTimer.StartTimer(Batch)
	require.Nil(t, err)
	time.Sleep(waitTime)
	err = batchTimer.RestartTimer(Batch)
	require.Nil(t, err)
	<-ch
	require.True(t, time.Since(start) > batchTimeout+waitTime)
}
