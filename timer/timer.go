package timer

import (
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/external"
)

func NewTimer(eventC chan interface{}, logger external.Logger) internal.Timer {
	return newTimerImpl(eventC, logger)
}

func (ti *timerImpl) StartTimer(name string, event interface{}) {
	ti.startTimer(name, event)
}

func (ti *timerImpl) StopTimer(name string) {
	ti.stopTimer(name)
}

func (ti *timerImpl) ClearTimer() {
	ti.clearTimer()
}
