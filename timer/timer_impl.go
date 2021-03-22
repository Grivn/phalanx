package timer

import (
	"strconv"
	"sync"
	"time"

	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/timer/types"
)

type titleTimer struct {
	timerName string        // the unique timer name
	timeout   time.Duration // default timeout of this timer
	isActive  sync.Map      // track all the timers with this timerName if it is active now
}

func (tt *titleTimer) store(key, value interface{}) {
	tt.isActive.Store(key, value)
}

func (tt *titleTimer) delete(key interface{}) {
	tt.isActive.Delete(key)
}

func (tt *titleTimer) has(key string) bool {
	_, ok := tt.isActive.Load(key)
	return ok
}

func (tt *titleTimer) count() int {
	length := 0
	tt.isActive.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

func (tt *titleTimer) clear() {
	tt.isActive.Range(func(key, _ interface{}) bool {
		tt.isActive.Delete(key)
		return true
	})
}

type timerImpl struct {
	tTimers   map[string]*titleTimer
	eventChan chan<- interface{}
	logger    external.Logger
}

func newTimerImpl(eventC chan interface{}, logger external.Logger) *timerImpl {
	tm := &timerImpl{
		tTimers:   make(map[string]*titleTimer),
		eventChan: eventC,
		logger:    logger,
	}

	return tm
}

func (tm *timerImpl) newTimer(name string, d time.Duration) {
	if d == 0 {
		switch name {
		case types.BinaryTagTimer:
			d = types.DefaultBinaryTagTimer
		}
	}

	tm.tTimers[name] = &titleTimer{
		timerName: name,
		timeout:   d,
	}
}

func (tm *timerImpl) Stop() {
	for timerName := range tm.tTimers {
		tm.stopTimer(timerName)
	}
}

func (tm *timerImpl) startTimer(name string, event types.TimeoutEvent) string {
	tm.stopTimer(name)

	timestamp := time.Now().UnixNano()
	key := strconv.FormatInt(timestamp, 10)
	tm.tTimers[name].store(key, true)

	send := func() {
		if tm.tTimers[name].has(key) {
			tm.eventChan <- event
		}
	}
	time.AfterFunc(tm.tTimers[name].timeout, send)
	return key
}

// stopTimer stops all timers with the same timerName.
func (tm *timerImpl) stopTimer(name string) {
	if !tm.containsTimer(name) {
		tm.logger.Errorf("Stop timer failed, timer %s not created yet!", name)
		return
	}

	tm.tTimers[name].clear()
}

// containsTimer returns true if there exists a timer named timerName
func (tm *timerImpl) containsTimer(timerName string) bool {
	_, ok := tm.tTimers[timerName]
	return ok
}

// getTimeoutValue gets the default timeout of the given timer
func (tm *timerImpl) getTimeoutValue(timerName string) time.Duration {
	if !tm.containsTimer(timerName) {
		tm.logger.Warningf("Get timeout failed!, timer %s not created yet!", timerName)
		return 0 * time.Second
	}
	return tm.tTimers[timerName].timeout
}

// setTimeoutValue sets the default timeout of the given timer with a new timeout
func (tm *timerImpl) setTimeoutValue(timerName string, d time.Duration) {
	if !tm.containsTimer(timerName) {
		tm.logger.Warningf("Set timeout failed!, timer %s not created yet!", timerName)
		return
	}
	tm.tTimers[timerName].timeout = d
}
