package utils

import (
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/external"
)

type localTimer struct {
	// author indicates current node identifier.
	author uint64

	// duration is the timeout interval for current timer.
	duration time.Duration

	// isActive indicates if current timer is active.
	isActive uint64

	// timeoutC is used to send timeout event.
	timeoutC chan<- bool

	// logger is used to print logs.
	logger external.Logger
}

func NewLocalTimer(author uint64, timeoutC chan bool, duration time.Duration, logger external.Logger) api.LocalTimer {
	return &localTimer{
		author:   author,
		duration: duration,
		timeoutC: timeoutC,
		logger:   logger,
	}
}

func (timer *localTimer) StartTimer() {
	timer.logger.Debugf("[%d] start partial order generation timer, duration %v", timer.author, timer.duration)
	atomic.StoreUint64(&timer.isActive, 1)

	f := func() {
		if atomic.LoadUint64(&timer.isActive) == 1 {
			timer.timeoutC <- true
		}
	}
	time.AfterFunc(timer.duration, f)
}

func (timer *localTimer) StopTimer() {
	timer.logger.Debugf("[%d] stop partial order generation timer", timer.author)
	atomic.StoreUint64(&timer.isActive, 0)
}
