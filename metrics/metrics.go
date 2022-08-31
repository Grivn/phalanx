package metrics

import (
	"github.com/Grivn/phalanx/common/types"
	"sync"
)

var mutex sync.Mutex

type Metrics struct {
	MetaPoolMetrics       *MetaPoolMetrics
	ExecutorMetrics       *ExecutorMetrics
	OrderRuleMetrics      *OrderRuleMetrics
	MediumTimeMetrics     *OrderRuleMetrics
	RuleCommitmentMetrics *RuleCommitmentMetrics
}

func NewMetrics() *Metrics {
	m := &Metrics{}
	m.MetaPoolMetrics = NewMetaPoolMetrics()
	m.ExecutorMetrics = NewExecutorMetrics()
	m.OrderRuleMetrics = NewOrderRuleMetrics()
	m.MediumTimeMetrics = NewOrderRuleMetrics()
	m.RuleCommitmentMetrics = NewRuleCommitmentMetrics()
	return m
}

// QueryMetrics returns metrics info.
func (ei *Metrics) QueryMetrics() types.MetricsInfo {
	return types.MetricsInfo{
		AvePackOrderLatency:      ei.MetaPoolMetrics.AvePackOrderLatency(),
		AveOrderLatency:          ei.MetaPoolMetrics.AveOrderLatency(),
		CurPackOrderLatency:      ei.MetaPoolMetrics.CurPackOrderLatency(),
		CurOrderLatency:          ei.MetaPoolMetrics.CurOrderLatency(),
		AveOrderSize:             ei.MetaPoolMetrics.AveOrderSize(),
		CommandPS:                ei.MetaPoolMetrics.CommandThroughput(),
		LogPS:                    ei.MetaPoolMetrics.LogThroughput(),
		GenLogPS:                 ei.MetaPoolMetrics.GenLogThroughput(),
		AveLogLatency:            ei.ExecutorMetrics.AveLogLatency(),
		CurLogLatency:            ei.ExecutorMetrics.CurLogLatency(),
		AveCommitStreamLatency:   ei.ExecutorMetrics.AveCommitStreamLatency(),
		CurCommitStreamLatency:   ei.ExecutorMetrics.CurCommitStreamLatency(),
		AveCommandInfoLatency:    ei.RuleCommitmentMetrics.AveCommandInfoLatency(),
		CurCommandInfoLatency:    ei.RuleCommitmentMetrics.CurCommandInfoLatency(),
		SafeCommandCount:         ei.OrderRuleMetrics.TotalSafeCommit,
		RiskCommandCount:         ei.OrderRuleMetrics.TotalRiskCommit,
		FrontAttackFromRisk:      ei.OrderRuleMetrics.FrontAttackFromRisk,
		FrontAttackFromSafe:      ei.OrderRuleMetrics.FrontAttackFromSafe,
		FrontAttackIntervalRisk:  ei.OrderRuleMetrics.FrontAttackIntervalRisk,
		FrontAttackIntervalSafe:  ei.OrderRuleMetrics.FrontAttackIntervalSafe,
		MSafeCommandCount:        ei.MediumTimeMetrics.TotalSafeCommit,
		MRiskCommandCount:        ei.MediumTimeMetrics.TotalRiskCommit,
		MFrontAttackFromRisk:     ei.MediumTimeMetrics.FrontAttackFromRisk,
		MFrontAttackFromSafe:     ei.MediumTimeMetrics.FrontAttackFromSafe,
		MFrontAttackIntervalRisk: ei.MediumTimeMetrics.FrontAttackIntervalRisk,
		MFrontAttackIntervalSafe: ei.MediumTimeMetrics.FrontAttackIntervalSafe,
	}
}
