package timer

import (
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/external"
)

type singleTimer struct {
	// duration is the timeout interval for current timer.
	duration time.Duration

	// isActive indicates if current timer is active.
	isActive uint64

	// timeoutC is used to send timeout event.
	timeoutC chan bool

	// logger is used to print logs.
	logger external.Logger
}

func NewSingleTimer(duration time.Duration, logger external.Logger) api.SingleTimer {
	return &singleTimer{
		duration: duration,
		timeoutC: make(chan bool),
		logger:   logger,
	}
}

func (timer *singleTimer) StartTimer() {
	timer.logger.Debugf("start timer, duration %v", timer.duration)
	atomic.StoreUint64(&timer.isActive, 1)

	f := func() {
		if atomic.LoadUint64(&timer.isActive) == 1 {
			timer.timeoutC <- true
		}
	}
	time.AfterFunc(timer.duration, f)
}

func (timer *singleTimer) StopTimer() {
	timer.logger.Debugf("stop timer")
	atomic.StoreUint64(&timer.isActive, 0)
}

func (timer *singleTimer) TimeoutChan() <-chan bool {
	return timer.timeoutC
}
