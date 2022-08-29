package metrics

import (
	"time"

	"github.com/Grivn/phalanx/common/types"
)

type RuleCommitmentMetrics struct {
	// TotalCommandInfo is the number of committed command info.
	TotalCommandInfo int

	//
	IntervalCommandInfo int

	// TotalLatency is the total latency since command info generation to be committed.
	TotalLatency int64

	//
	IntervalLatency int64
}

func NewRuleCommitmentMetrics() *RuleCommitmentMetrics {
	return &RuleCommitmentMetrics{}
}

func (m *RuleCommitmentMetrics) CommitFrontCommandInfo(frontC *types.CommandInfo) {
	sub := time.Now().UnixNano() - frontC.GTime
	m.TotalCommandInfo++
	m.TotalLatency += sub
	m.IntervalCommandInfo++
	m.IntervalLatency += sub
}

func (m *RuleCommitmentMetrics) AveCommandInfoLatency() float64 {
	// aveCommandInfoLatency returns average latency of command info to be committed.
	if m.TotalCommandInfo == 0 {
		return 0
	}
	return types.NanoToMillisecond(m.TotalLatency / int64(m.TotalCommandInfo))
}

func (m *RuleCommitmentMetrics) CurCommandInfoLatency() float64 {
	// curCommandInfoLatency returns average latency of command info to be committed.
	if m.IntervalCommandInfo == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(m.IntervalLatency / int64(m.IntervalCommandInfo))
	m.IntervalLatency = 0
	m.IntervalCommandInfo = 0
	return ret
}
