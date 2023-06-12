package metrics

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	types "github.com/Grivn/phalanx/pkg/common/types"
	"sync"
	"time"
)

type MetaPoolMetrics struct {
	// mutex is used to process concurrency problem for this metrics instance.
	mutex sync.Mutex

	// TotalCommands is the number of commands selected into partial order.
	TotalCommands int

	//
	IntervalCommands int

	// TotalSelectLatency is the total latency of interval since receive command to generate pre-order.
	TotalSelectLatency int64

	//
	IntervalSelectLatency int64

	// totalOrderLogs is the number of order logs.
	TotalOrderLogs int

	//
	IntervalOrderLogs int

	// totalLatency is the total latency of the generation of trusted order logs.
	TotalLatency int64

	//
	IntervalLatency int64

	//
	OrderCount int

	//
	OrderSize int

	//
	CommandCount int

	//
	StartTime int64

	//
	GenOrder int
}

func NewMetaPoolMetrics() *MetaPoolMetrics {
	return &MetaPoolMetrics{}
}

func (m *MetaPoolMetrics) ProcessCommand() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.CommandCount == 0 {
		m.StartTime = time.Now().UnixNano()
	}
	m.CommandCount++
}

func (m *MetaPoolMetrics) SelectCommand(cIndex *types.CommandIndex) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nowT := time.Now().UnixNano()
	m.TotalCommands++
	m.TotalSelectLatency += nowT - cIndex.RTime

	m.IntervalCommands++
	m.IntervalSelectLatency += nowT - cIndex.RTime
}

func (m *MetaPoolMetrics) GenerateOrder() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.GenOrder++
}

func (m *MetaPoolMetrics) PartialOrderQuorum(pOrder *protos.PartialOrder) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// collect metrics.
	m.TotalOrderLogs++
	m.TotalLatency += pOrder.OrderedTime - pOrder.TimestampList()[0]

	m.IntervalOrderLogs++
	m.IntervalLatency += pOrder.OrderedTime - pOrder.TimestampList()[0]

	m.OrderCount++
	m.OrderSize += len(pOrder.PreOrder.CommandList)
}

func (m *MetaPoolMetrics) AvePackOrderLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.TotalCommands == 0 {
		return 0
	}
	return types.NanoToMillisecond(m.TotalSelectLatency / int64(m.TotalCommands))
}

func (m *MetaPoolMetrics) CurPackOrderLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.IntervalCommands == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(m.IntervalSelectLatency / int64(m.IntervalCommands))
	m.IntervalCommands = 0
	m.IntervalSelectLatency = 0
	return ret
}

func (m *MetaPoolMetrics) AveOrderLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.TotalOrderLogs == 0 {
		return 0
	}
	return types.NanoToMillisecond(m.TotalLatency / int64(m.TotalOrderLogs))
}

func (m *MetaPoolMetrics) CurOrderLatency() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.IntervalOrderLogs == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(m.IntervalLatency / int64(m.IntervalOrderLogs))
	m.IntervalOrderLogs = 0
	m.IntervalLatency = 0
	return ret
}

func (m *MetaPoolMetrics) AveOrderSize() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.OrderCount == 0 {
		return 0
	}
	return m.OrderSize / m.OrderCount
}

func (m *MetaPoolMetrics) CommandThroughput() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	interval := types.NanoToMillisecond(time.Now().UnixNano() - m.StartTime)
	fInterval := interval / 1000
	return float64(m.CommandCount) / fInterval
}

func (m *MetaPoolMetrics) LogThroughput() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	interval := types.NanoToMillisecond(time.Now().UnixNano() - m.StartTime)
	fInterval := interval / 1000
	return float64(m.OrderCount) / fInterval
}

func (m *MetaPoolMetrics) GenLogThroughput() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	interval := types.NanoToMillisecond(time.Now().UnixNano() - m.StartTime)
	fInterval := interval / 1000
	return float64(m.GenOrder) / fInterval
}
