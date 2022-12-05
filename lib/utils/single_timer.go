package utils

import (
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/external"
)

type singleTimer struct {
	// duration is the timeout interval for current timer.
	duration time.Duration

	// isActive indicates if current timer is active.
	isActive uint64

	// timeoutC is used to send timeout event.
	timeoutC chan<- bool

	// logger is used to print logs.
	logger external.Logger
}

func NewSingleTimer(timeoutC chan bool, duration time.Duration, logger external.Logger) api.SingleTimer {
	return &singleTimer{
		duration: duration,
		timeoutC: timeoutC,
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
