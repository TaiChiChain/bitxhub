package timer

import (
	"fmt"
	"strconv"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/sirupsen/logrus"
)

const (
	Batch             TimeoutEvent = "Batch"
	NoTxBatch         TimeoutEvent = "NoTxBatch"
	RemoveTx          TimeoutEvent = "RemoveTx"
	CleanEmptyAccount TimeoutEvent = "CleanEmptyAccount"
)

type TimeoutEvent string

type Timer interface {
	CreateTimer(name TimeoutEvent, d time.Duration, handler func(name TimeoutEvent)) error
	StartTimer(name TimeoutEvent) error

	RestartTimer(name TimeoutEvent) error

	StopTimer(name TimeoutEvent)

	IsTimerActive(name TimeoutEvent) bool

	Stop()
}

// singleTimer manages timer with the same timer name, which, we allow different timer with the same timer name, such as:
// we allow several request timers at the same time, each timer started after received a new request batch
type singleTimer struct {
	timerName TimeoutEvent                            // the unique timer name
	timeout   time.Duration                           // default timeout of this timer
	isActive  cmap.ConcurrentMap[string, *time.Timer] // track all the timers with this timerName if it is active now
	handler   func(name TimeoutEvent)
}

func (tt *singleTimer) clear() {
	for _, timer := range tt.isActive.Items() {
		timer.Stop()
	}
	tt.isActive.Clear()
}

func (tt *singleTimer) isTimerActive() bool {
	return tt.isActive.Count() > 0
}

// TimerManager manages consensus used timers.
type TimerManager struct {
	timersM map[TimeoutEvent]*singleTimer
	logger  logrus.FieldLogger
}

// NewTimerManager news a timer with default timeout.
func NewTimerManager(logger logrus.FieldLogger) *TimerManager {
	return &TimerManager{
		timersM: make(map[TimeoutEvent]*singleTimer),
		logger:  logger,
	}
}

func (tm *TimerManager) CreateTimer(name TimeoutEvent, d time.Duration, handler func(name TimeoutEvent)) error {
	if d == 0 {
		return fmt.Errorf("invalid timeout %v", d)
	}
	tm.timersM[name] = &singleTimer{
		timerName: name,
		isActive:  cmap.New[*time.Timer](),
		timeout:   d,
		handler:   handler,
	}
	return nil
}

// Stop stops all timers managed by TimerManager
func (tm *TimerManager) Stop() {
	for timerName := range tm.timersM {
		tm.stopTimer(timerName)
	}
}

func (tm *TimerManager) StopTimer(timerName TimeoutEvent) {
	tm.stopTimer(timerName)
}

// stopTimer stops all timers with the same timerName.
func (tm *TimerManager) stopTimer(timerName TimeoutEvent) {
	if tm.containsTimer(timerName) {
		tm.timersM[timerName].clear()
		tm.logger.Debugf("Timer %s stopped", timerName)
	}
}

// containsTimer returns true if there exists a timer named timerName
func (tm *TimerManager) containsTimer(timerName TimeoutEvent) bool {
	if t, ok := tm.timersM[timerName]; ok {
		return t.isTimerActive()
	}
	return false
}

// StartTimer starts the timer with the given name and default timeout, then sets the event which will be triggered
// after this timeout
func (tm *TimerManager) StartTimer(name TimeoutEvent) error {
	if _, ok := tm.timersM[name]; !ok {
		return fmt.Errorf("timer %s doesn't exist", name)
	}
	tm.stopTimer(name)

	timestamp := time.Now().UnixNano()
	key := strconv.FormatInt(timestamp, 10)
	fn := func() {
		if tm.timersM[name].isActive.Has(key) {
			tm.timersM[name].handler(name)
		}
	}
	afterTimer := time.AfterFunc(tm.timersM[name].timeout, fn)
	tm.timersM[name].isActive.Set(key, afterTimer)
	tm.logger.Debugf("Timer %s started", name)
	return nil
}

func (tm *TimerManager) RestartTimer(name TimeoutEvent) error {
	if _, ok := tm.timersM[name]; !ok {
		return fmt.Errorf("timer %s doesn't exist", name)
	}
	tm.stopTimer(name)
	timestamp := time.Now().UnixNano()
	key := strconv.FormatInt(timestamp, 10)
	fn := func() {
		if tm.timersM[name].isActive.Has(key) {
			tm.timersM[name].handler(name)
		}
	}
	afterTimer := time.AfterFunc(tm.timersM[name].timeout, fn)
	tm.timersM[name].isActive.Set(key, afterTimer)
	tm.logger.Debugf("restart Timer %s", name)
	return nil
}

func (tm *TimerManager) IsTimerActive(name TimeoutEvent) bool {
	return tm.isTimerActive(name)
}

func (tm *TimerManager) isTimerActive(name TimeoutEvent) bool {
	if t, ok := tm.timersM[name]; ok {
		return t.isTimerActive()
	}
	return false
}
