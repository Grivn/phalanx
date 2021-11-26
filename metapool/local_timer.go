package metapool

import (
	"sync/atomic"
	"time"
)

type localTimer struct {
	// duration is the timeout interval for current timer.
	duration time.Duration

	// isActive indicates if current timer is active.
	isActive uint64

	// timeoutC is used to send timeout event.
	timeoutC chan<- bool
}

func newLocalTimer(timeoutC chan bool, duration time.Duration) *localTimer {
	return &localTimer{
		duration: duration,
		timeoutC: timeoutC,
	}
}

// startTimer starts current timer.
func (timer *localTimer) startTimer() {
	atomic.StoreUint64(&timer.isActive, 1)

	f := func() {
		if atomic.LoadUint64(&timer.isActive) == 1 {
			timer.timeoutC <- true
		}
	}
	time.AfterFunc(timer.duration, f)
}

// stopTimer stops current timer.
func (timer *localTimer) stopTimer() {
	atomic.StoreUint64(&timer.isActive, 0)
}
