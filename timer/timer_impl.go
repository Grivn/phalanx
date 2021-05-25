package timer

import (
	types2 "github.com/Grivn/phalanx/common/types"
	"strconv"
	"sync"
	"time"

	"github.com/Grivn/phalanx/external"
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

	tm.newTimer(types2.BinaryTagTimer, types2.DefaultBinaryTagTimer)
	tm.newTimer(types2.TxPoolTimer, types2.DefaultTxPoolTimer)

	return tm
}

func (ti *timerImpl) newTimer(name string, d time.Duration) {
	if d == 0 {
		ti.logger.Warningf("time duration is 0")
		return
	}

	ti.tTimers[name] = &titleTimer{
		timerName: name,
		timeout:   d,
	}
}

func (ti *timerImpl) clearTimer() {
	for timerName := range ti.tTimers {
		ti.stopTimer(timerName)
	}
}

func (ti *timerImpl) startTimer(name string, event interface{}) string {
	ti.stopTimer(name)

	timestamp := time.Now().UnixNano()
	key := strconv.FormatInt(timestamp, 10)
	ti.tTimers[name].store(key, true)

	send := func() {
		if ti.tTimers[name].has(key) {
			ti.eventChan <- event
		}
	}
	time.AfterFunc(ti.tTimers[name].timeout, send)
	return key
}

// stopTimer stops all timers with the same timerName.
func (ti *timerImpl) stopTimer(name string) {
	if !ti.containsTimer(name) {
		ti.logger.Errorf("Stop timer failed, timer %s not created yet!", name)
		return
	}

	ti.tTimers[name].clear()
}

// containsTimer returns true if there exists a timer named timerName
func (ti *timerImpl) containsTimer(timerName string) bool {
	_, ok := ti.tTimers[timerName]
	return ok
}

// getTimeoutValue gets the default timeout of the given timer
func (ti *timerImpl) getTimeoutValue(timerName string) time.Duration {
	if !ti.containsTimer(timerName) {
		ti.logger.Warningf("Get timeout failed!, timer %s not created yet!", timerName)
		return 0 * time.Second
	}
	return ti.tTimers[timerName].timeout
}

// setTimeoutValue sets the default timeout of the given timer with a new timeout
func (ti *timerImpl) setTimeoutValue(timerName string, d time.Duration) {
	if !ti.containsTimer(timerName) {
		ti.logger.Warningf("Set timeout failed!, timer %s not created yet!", timerName)
		return
	}
	ti.tTimers[timerName].timeout = d
}
