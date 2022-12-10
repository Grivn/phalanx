package metrics

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"sync"
	"time"
)

type ExecutorMetrics struct {
	// mutex is used to process concurrency problem for this metrics instance.
	mutex sync.Mutex

	//============================= metrics =================================

	// TotalLogs tracks the number of committed partial order logs.
	TotalLogs int

	//
	IntervalLogs int

	// TotalLatency tracks the total latency since partial order generation to commitment.
	TotalLatency int64

	//
	IntervalLatency int64

	//
	TotalStreams int

	//
	IntervalStreams int

	//
	TotalCommitStreamLatency int64

	//
	IntervalCommitStreamLatency int64
}

func NewExecutorMetrics() *ExecutorMetrics {
	return &ExecutorMetrics{}
}

func (m *ExecutorMetrics) CommitPartialOrder(pOrder *protos.PartialOrder) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub := time.Now().UnixNano() - pOrder.OrderedTime

	// collect order log metrics.
	m.TotalLogs++
	m.TotalLatency += sub

	m.IntervalLogs++
	m.IntervalLatency += sub
}

func (m *ExecutorMetrics) CommitStream(start time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub := time.Now().Sub(start).Milliseconds()
	m.TotalCommitStreamLatency += sub
	m.TotalStreams++
	m.IntervalCommitStreamLatency += sub
	m.IntervalStreams++
}

func (m *ExecutorMetrics) AveLogLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.TotalLogs == 0 {
		return 0
	}
	return types.NanoToMillisecond(m.TotalLatency / int64(m.TotalLogs))
}

func (m *ExecutorMetrics) CurLogLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// returns average latency of partial orders to be committed.
	if m.IntervalLogs == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(m.IntervalLatency / int64(m.IntervalLogs))
	m.IntervalLatency = 0
	m.IntervalLogs = 0
	return ret
}

func (m *ExecutorMetrics) AveCommitStreamLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// AveCommitStreamLatency returns average latency of commitment of query stream.
	if m.TotalStreams == 0 {
		return 0
	}
	return types.NanoToMillisecond(m.TotalCommitStreamLatency / int64(m.TotalStreams))
}

func (m *ExecutorMetrics) CurCommitStreamLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// curCommitStreamLatency returns average latency of commitment of query stream.
	if m.IntervalStreams == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(m.IntervalCommitStreamLatency / int64(m.IntervalStreams))
	m.IntervalCommitStreamLatency = 0
	m.IntervalStreams = 0
	return ret
}
