package metrics

import (
	"github.com/Grivn/phalanx/common/types"
)

type Metrics struct {
	MetaPoolMetrics       *MetaPoolMetrics
	ExecutorMetrics       *ExecutorMetrics
	OrderRuleMetrics      *OrderRuleMetrics
	MediumTimeMetrics     *OrderRuleMetrics
	RuleCommitmentMetrics *RuleCommitmentMetrics
}

func NewMetrics() *Metrics {
	return &Metrics{
		MetaPoolMetrics:       NewMetaPoolMetrics(),
		ExecutorMetrics:       NewExecutorMetrics(),
		OrderRuleMetrics:      NewOrderRuleMetrics(),
		MediumTimeMetrics:     NewOrderRuleMetrics(),
		RuleCommitmentMetrics: NewRuleCommitmentMetrics(),
	}
}

// QueryMetrics returns metrics info.
func (ei *Metrics) QueryMetrics() types.MetricsInfo {
	phalanxOrder := ei.OrderRuleMetrics.QueryMetrics()
	mediumTOrder := ei.MediumTimeMetrics.QueryMetrics()
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
		SafeCommandCount:         phalanxOrder.SafeCommandCount,
		RiskCommandCount:         phalanxOrder.RiskCommandCount,
		FrontAttackFromRisk:      phalanxOrder.FrontAttackFromRisk,
		FrontAttackFromSafe:      phalanxOrder.FrontAttackFromSafe,
		FrontAttackIntervalRisk:  phalanxOrder.FrontAttackIntervalRisk,
		FrontAttackIntervalSafe:  phalanxOrder.FrontAttackIntervalSafe,
		SuccessRates:             phalanxOrder.SuccessRates,
		MSafeCommandCount:        mediumTOrder.SafeCommandCount,
		MRiskCommandCount:        mediumTOrder.RiskCommandCount,
		MFrontAttackFromRisk:     mediumTOrder.FrontAttackFromRisk,
		MFrontAttackFromSafe:     mediumTOrder.FrontAttackFromSafe,
		MFrontAttackIntervalRisk: mediumTOrder.FrontAttackIntervalRisk,
		MFrontAttackIntervalSafe: mediumTOrder.FrontAttackIntervalSafe,
		MSuccessRates:            mediumTOrder.SuccessRates,
	}
}
