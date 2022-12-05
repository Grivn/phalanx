package utils

import (
	"sync"
	"time"
)

// titledTimer manages timer with the same timer name, which, we allow different timer with the same timer name, such as:
// we allow several request timers at the same time, each timer started after received a new request batch
type titledTimer struct {
	timerName string        // the unique timer name
	timeout   time.Duration // default timeout of this timer
	isActive  sync.Map      // track all the timers with this timerName if it is active now
}

func (tt *titledTimer) store(key, value interface{}) {
	tt.isActive.Store(key, value)
}

func (tt *titledTimer) delete(key interface{}) {
	tt.isActive.Delete(key)
}

func (tt *titledTimer) has(key string) bool {
	_, ok := tt.isActive.Load(key)
	return ok
}

func (tt *titledTimer) count() int {
	length := 0
	tt.isActive.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

func (tt *titledTimer) clear() {
	tt.isActive.Range(func(key, _ interface{}) bool {
		tt.isActive.Delete(key)
		return true
	})
}
