package api

import (
	"time"

	"github.com/Grivn/phalanx/common/types"
)

type SingleTimer interface {
	// StartTimer starts current timer.
	StartTimer()

	// StopTimer stops current timer.
	StopTimer()
}

type TitledTimerManager interface {
	// CreateTimer creates a titled timer with the given name and timeout duration, then record it in timer manager.
	CreateTimer(name string, d time.Duration)

	// StartTimer starts the timer with the given name and default timeout,
	// then sets the event which will be triggered after this timeout duration.
	StartTimer(name string, event types.LocalEvent) string

	// StopTimer stops all timers with the same name.
	StopTimer(name string)

	// StopOneTimer stops one timer by the name and index.
	StopOneTimer(name string, key string)

	// Stop stops all the current titled timers.
	Stop()

	// GetTimeoutValue gets the default timeout of the given timer
	GetTimeoutValue(name string) time.Duration

	// SetTimeoutValue sets the default timeout of the given timer with a new timeout
	SetTimeoutValue(name string, d time.Duration)
}
