package txmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"sync/atomic"
	"time"
)

type localTimer struct {
	// author indicates current node identifier.
	author uint64

	// duration is the timeout interval for current timer.
	duration time.Duration

	// isActive indicates if current timer is active.
	isActive uint64

	// timeoutC is used to send timeout event.
	timeoutC chan<- *protos.Command

	// logger is used to print logs.
	logger external.Logger
}

func newLocalTimer(author uint64, timeoutC chan *protos.Command, duration time.Duration, logger external.Logger) *localTimer {
	return &localTimer{
		author:   author,
		duration: duration,
		timeoutC: timeoutC,
		logger:   logger,
	}
}

// startTimer starts current timer.
func (timer *localTimer) startTimerOnlyOne(command *protos.Command) {
	if atomic.LoadUint64(&timer.isActive) == 1 {
		return
	}

	timer.logger.Debugf("[%d] start partial order generation timer, duration %v", timer.author, timer.duration)
	atomic.StoreUint64(&timer.isActive, 1)

	f := func() {
		if atomic.LoadUint64(&timer.isActive) == 1 {
			timer.timeoutC <- command
		}
	}
	time.AfterFunc(timer.duration, f)
}

// stopTimer stops current timer.
func (timer *localTimer) stopTimer() {
	timer.logger.Debugf("[%d] stop partial order generation timer", timer.author)
	atomic.StoreUint64(&timer.isActive, 0)
}
