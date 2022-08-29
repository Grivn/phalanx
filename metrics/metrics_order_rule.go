package metrics

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type OrderRuleMetrics struct {
	// TotalSafeCommit tracks the number of command committed from safe path.
	TotalSafeCommit int

	// TotalRiskCommit tracks the number of command committed from risk path.
	TotalRiskCommit int

	//======================================== detect attack info =======================================================

	// CommandRecorder key proposer id value latest committed seq, in order to detect front attacks.
	CommandRecorder map[uint64]uint64

	// FrontAttackFromSafe is used to record the front attacked command request with safe front set.
	FrontAttackFromSafe int

	// FrontAttackFromRisk is used to record the front attacked command request with risk front set.
	FrontAttackFromRisk int

	// FrontAttackIntervalSafe is used to record the front attacked command request with safe of interval relationship.
	FrontAttackIntervalSafe int

	// FrontAttackIntervalRisk is used to record the front attacked command request with risk of interval relationship.
	FrontAttackIntervalRisk int
}

func NewOrderRuleMetrics() *OrderRuleMetrics {
	return &OrderRuleMetrics{CommandRecorder: make(map[uint64]uint64)}
}

func (m *OrderRuleMetrics) CommitBlock(blk types.InnerBlock) {
	m.DetectFrontSetTypes(!blk.Safe)
	m.DetectFrontAttackGivenRelationship(!blk.Safe, blk.Command)
	m.DetectFrontAttackIntervalRelationship(!blk.Safe, blk.Command)
	m.UpdateFrontAttackDetector(blk.Command)
}

func (m *OrderRuleMetrics) DetectFrontSetTypes(risk bool) {
	if !risk {
		m.TotalSafeCommit++
	} else {
		m.TotalRiskCommit++
	}
}

func (m *OrderRuleMetrics) DetectFrontAttackGivenRelationship(risk bool, command *protos.Command) {
	// detect the front attack towards given relationship.
	current := m.CommandRecorder[command.Author]

	if command.Sequence != current+1 {
		if risk {
			m.FrontAttackFromRisk++
		} else {
			m.FrontAttackFromSafe++
		}
	}
}

func (m *OrderRuleMetrics) DetectFrontAttackIntervalRelationship(risk bool, command *protos.Command) {
	// detect the front attack towards interval relationship.
	if command.FrontRunner == nil {
		return
	}

	if command.FrontRunner.Sequence > m.CommandRecorder[command.FrontRunner.Author] {
		if risk {
			m.FrontAttackFromRisk++
			m.FrontAttackIntervalRisk++
		} else {
			m.FrontAttackFromSafe++
			m.FrontAttackIntervalSafe++
		}
	}
}

func (m *OrderRuleMetrics) UpdateFrontAttackDetector(command *protos.Command) {
	// update the detector for front attacked command requests.
	if command.Sequence > m.CommandRecorder[command.Author] {
		m.CommandRecorder[command.Author] = command.Sequence
	}
}
