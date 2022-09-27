package metrics

import (
	"sync"
	"time"

	"github.com/Grivn/phalanx/common/types"
)

type CommitmentMetrics struct {
	// mutex is used to process concurrency problem for this metrics instance.
	mutex sync.Mutex

	// TotalCommandInfo is the number of committed command info.
	TotalCommandInfo int

	//
	IntervalCommandInfo int

	// TotalLatency is the total latency since command info generation to be committed.
	TotalLatency int64

	//
	IntervalLatency int64
}

func NewCommitmentMetrics() *CommitmentMetrics {
	return &CommitmentMetrics{}
}

func (m *CommitmentMetrics) CommitFrontCommandInfo(frontC *types.CommandInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub := time.Now().UnixNano() - frontC.GTime
	m.TotalCommandInfo++
	m.TotalLatency += sub
	m.IntervalCommandInfo++
	m.IntervalLatency += sub
}

func (m *CommitmentMetrics) AveCommandInfoLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// aveCommandInfoLatency returns average latency of command info to be committed.
	if m.TotalCommandInfo == 0 {
		return 0
	}
	return types.NanoToMillisecond(m.TotalLatency / int64(m.TotalCommandInfo))
}

func (m *CommitmentMetrics) CurCommandInfoLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// curCommandInfoLatency returns average latency of command info to be committed.
	if m.IntervalCommandInfo == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(m.IntervalLatency / int64(m.IntervalCommandInfo))
	m.IntervalLatency = 0
	m.IntervalCommandInfo = 0
	return ret
}
