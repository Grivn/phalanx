package execsimple

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sort"
)

type orderMediumT struct {
	seqNo uint64

	// logger is used to print logs.
	logger external.Logger

	// totalSafeCommit tracks the number of command committed from safe path.
	totalSafeCommit int

	// totalRiskCommit tracks the number of command committed from risk path.
	totalRiskCommit int

	//======================================== detect attack info =======================================================

	// commandRecorder key proposer id value latest committed seq, in order to detect front attacks.
	commandRecorder map[uint64]uint64

	// frontAttackFromSafe is used to record the front attacked command request with safe front set.
	frontAttackFromSafe int

	// frontAttackFromRisk is used to record the front attacked command request with risk front set.
	frontAttackFromRisk int

	// frontAttackIntervalSafe is used to record the front attacked command request with safe of interval relationship.
	frontAttackIntervalSafe int

	// frontAttackIntervalRisk is used to record the front attacked command request with risk of interval relationship.
	frontAttackIntervalRisk int

	blocks types.SortableInnerBlocks
}

func newOrderMediumT(logger external.Logger) *orderMediumT {
	return &orderMediumT{
		blocks:          nil,
		commandRecorder: make(map[uint64]uint64),
		logger:          logger,
	}
}

func (ot *orderMediumT) commitAccordingMediumT(block types.InnerBlock, seqNo uint64) {
	block.Timestamp = block.MediumT
	ot.blocks = append(ot.blocks, block)

	if len(ot.blocks) < 1000 {
		return
	}
	sort.Sort(ot.blocks)
	commitBlocks := ot.blocks[:10]
	updateBlocks := types.SortableInnerBlocks{}
	for index, blk := range ot.blocks {
		if index < 10 {
			continue
		}
		updateBlocks = append(updateBlocks, blk)
	}
	ot.blocks = updateBlocks

	for _, blk := range commitBlocks {
		ot.detectFrontSetTypes(!blk.Safe)
		ot.detectFrontAttackGivenRelationship(!blk.Safe, blk.Command)
		ot.detectFrontAttackIntervalRelationship(!blk.Safe, blk.Command)
		ot.updateFrontAttackDetector(blk.Command)
	}
}

func (ot *orderMediumT) detectFrontSetTypes(risk bool) {
	if !risk {
		ot.totalSafeCommit++
	} else {
		ot.totalRiskCommit++
	}
}

func (ot *orderMediumT) detectFrontAttackGivenRelationship(risk bool, command *protos.Command) {
	// detect the front attack towards given relationship.
	current := ot.commandRecorder[command.Author]

	if command.Sequence != current+1 {
		if risk {
			ot.frontAttackFromRisk++
		} else {
			ot.frontAttackFromSafe++
		}
	}
}

func (ot *orderMediumT) detectFrontAttackIntervalRelationship(risk bool, command *protos.Command) {
	// detect the front attack towards interval relationship.
	if command.FrontRunner == nil {
		return
	}

	if command.FrontRunner.Sequence > ot.commandRecorder[command.FrontRunner.Author] {
		if risk {
			ot.frontAttackFromRisk++
			ot.frontAttackIntervalRisk++
		} else {
			ot.frontAttackFromSafe++
			ot.frontAttackIntervalSafe++
		}
	}
}

func (ot *orderMediumT) updateFrontAttackDetector(command *protos.Command) {
	// update the detector for front attacked command requests.
	if command.Sequence > ot.commandRecorder[command.Author] {
		ot.commandRecorder[command.Author] = command.Sequence
	}
}
