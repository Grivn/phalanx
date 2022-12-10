package timer

import (
	"strconv"
	"time"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
)

// titledTimerManager manages common used timers.
type titledTimerManager struct {
	tTimers map[string]*titledTimer
	eventC  chan interface{}
	logger  external.Logger
}

func NewTitledTimerManager(logger external.Logger) api.TitledTimerManager {
	tm := &titledTimerManager{
		tTimers: make(map[string]*titledTimer),
		eventC:  make(chan interface{}),
		logger:  logger,
	}
	return tm
}

func (ttm *titledTimerManager) CreateTimer(name string, d time.Duration) {
	ttm.tTimers[name] = &titledTimer{
		timerName: name,
		timeout:   d,
	}
}

// Stop stops all timers managed by titledTimerManager
func (ttm *titledTimerManager) Stop() {
	for name := range ttm.tTimers {
		ttm.StopTimer(name)
	}
}

func (ttm *titledTimerManager) StartTimer(name string, event types.LocalEvent) string {
	ttm.StopTimer(name)

	timestamp := time.Now().UnixNano()
	key := strconv.FormatInt(timestamp, 10)
	ttm.tTimers[name].store(key, true)

	send := func() {
		if ttm.tTimers[name].has(key) {
			ttm.eventC <- event
		}
	}
	time.AfterFunc(ttm.tTimers[name].timeout, send)
	return key
}

func (ttm *titledTimerManager) EventChan() <-chan interface{} {
	return ttm.eventC
}

func (ttm *titledTimerManager) StopTimer(name string) {
	if !ttm.containsTimer(name) {
		ttm.logger.Errorf("Stop timer failed, timer %s not created yet!", name)
		return
	}

	ttm.tTimers[name].clear()
}

func (ttm *titledTimerManager) StopOneTimer(name string, key string) {
	if !ttm.containsTimer(name) {
		ttm.logger.Errorf("Stop timer failed!, timer %s not created yet!", name)
		return
	}
	ttm.tTimers[name].delete(key)
}

func (ttm *titledTimerManager) GetTimeoutValue(name string) time.Duration {
	if !ttm.containsTimer(name) {
		ttm.logger.Errorf("Get timeout failed!, timer %s not created yet!", name)
		return 0 * time.Second
	}
	return ttm.tTimers[name].timeout
}

func (ttm *titledTimerManager) SetTimeoutValue(name string, d time.Duration) {
	if !ttm.containsTimer(name) {
		ttm.logger.Errorf("Set timeout failed!, timer %s not created yet!", name)
		return
	}
	ttm.tTimers[name].timeout = d
}

// containsTimer returns true if there exists a timer named name
func (ttm *titledTimerManager) containsTimer(name string) bool {
	_, ok := ttm.tTimers[name]
	return ok
}