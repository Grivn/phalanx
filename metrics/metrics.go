package metrics

import (
	"github.com/Grivn/phalanx/common/types"
)

type Metrics struct {
	MetaPoolMetrics        *MetaPoolMetrics
	ExecutorMetrics        *ExecutorMetrics
	PhalanxAnchorMetrics   *ManipulationMetrics
	TimestampAnchorMetrics *ManipulationMetrics
	TimestampBasedMetrics  *ManipulationMetrics
	CommitmentMetrics      *CommitmentMetrics
}

func NewMetrics() *Metrics {
	return &Metrics{
		MetaPoolMetrics:        NewMetaPoolMetrics(),
		ExecutorMetrics:        NewExecutorMetrics(),
		PhalanxAnchorMetrics:   NewManipulationMetrics(),
		TimestampAnchorMetrics: NewManipulationMetrics(),
		TimestampBasedMetrics:  NewManipulationMetrics(),
		CommitmentMetrics:      NewCommitmentMetrics(),
	}
}

// QueryMetrics returns metrics info.
func (ei *Metrics) QueryMetrics() types.MetricsInfo {
	phalanxOrder := ei.PhalanxAnchorMetrics.QueryMetrics()
	mediumTOrder := ei.TimestampBasedMetrics.QueryMetrics()
	timeAnchorOrder := ei.TimestampAnchorMetrics.QueryMetrics()
	return types.MetricsInfo{
		AvePackOrderLatency:       ei.MetaPoolMetrics.AvePackOrderLatency(),
		AveOrderLatency:           ei.MetaPoolMetrics.AveOrderLatency(),
		CurPackOrderLatency:       ei.MetaPoolMetrics.CurPackOrderLatency(),
		CurOrderLatency:           ei.MetaPoolMetrics.CurOrderLatency(),
		AveOrderSize:              ei.MetaPoolMetrics.AveOrderSize(),
		CommandPS:                 ei.MetaPoolMetrics.CommandThroughput(),
		LogPS:                     ei.MetaPoolMetrics.LogThroughput(),
		GenLogPS:                  ei.MetaPoolMetrics.GenLogThroughput(),
		AveLogLatency:             ei.ExecutorMetrics.AveLogLatency(),
		CurLogLatency:             ei.ExecutorMetrics.CurLogLatency(),
		AveCommitStreamLatency:    ei.ExecutorMetrics.AveCommitStreamLatency(),
		CurCommitStreamLatency:    ei.ExecutorMetrics.CurCommitStreamLatency(),
		AveCommandInfoLatency:     ei.CommitmentMetrics.AveCommandInfoLatency(),
		CurCommandInfoLatency:     ei.CommitmentMetrics.CurCommandInfoLatency(),
		SafeCommandCount:          phalanxOrder.SafeCommandCount,
		RiskCommandCount:          phalanxOrder.RiskCommandCount,
		FrontAttackFromRisk:       phalanxOrder.FrontAttackFromRisk,
		FrontAttackFromSafe:       phalanxOrder.FrontAttackFromSafe,
		FrontAttackIntervalRisk:   phalanxOrder.FrontAttackIntervalRisk,
		FrontAttackIntervalSafe:   phalanxOrder.FrontAttackIntervalSafe,
		SuccessRates:              phalanxOrder.SuccessRates,
		MSafeCommandCount:         mediumTOrder.SafeCommandCount,
		MRiskCommandCount:         mediumTOrder.RiskCommandCount,
		MFrontAttackFromRisk:      mediumTOrder.FrontAttackFromRisk,
		MFrontAttackFromSafe:      mediumTOrder.FrontAttackFromSafe,
		MFrontAttackIntervalRisk:  mediumTOrder.FrontAttackIntervalRisk,
		MFrontAttackIntervalSafe:  mediumTOrder.FrontAttackIntervalSafe,
		MSuccessRates:             mediumTOrder.SuccessRates,
		TASafeCommandCount:        timeAnchorOrder.SafeCommandCount,
		TARiskCommandCount:        timeAnchorOrder.RiskCommandCount,
		TAFrontAttackFromRisk:     timeAnchorOrder.FrontAttackFromRisk,
		TAFrontAttackFromSafe:     timeAnchorOrder.FrontAttackFromSafe,
		TAFrontAttackIntervalRisk: timeAnchorOrder.FrontAttackIntervalRisk,
		TAFrontAttackIntervalSafe: timeAnchorOrder.FrontAttackIntervalSafe,
		TASuccessRates:            timeAnchorOrder.SuccessRates,
	}
}
