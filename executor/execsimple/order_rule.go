package execsimple

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

type orderRule struct {
	//============================== basic info =====================================

	// author indicates the identifier of current node.
	author uint64

	// seqNo indicates the order of inner blocks.
	seqNo uint64

	//======================== order rules ======================================

	// collect indicates the collection rule of phalanx:
	// which partial orders would be selected into execution process to compare order.
	collect *collectionRule

	// execute indicates the execution rule of phalanx:
	// which commands would be selected into commitment process to generate blocks.
	// here, we should take 'Natural Order' into thought.
	execute *executionRule

	// commit indicates the commitment rule of phalanx:
	// generate blocks and assign sequence order for them.
	// here, the block generation would follow the 'Free Will' of participants.
	commit *commitmentRule

	//============================= internal interfaces =========================================

	// reload is used to notify client instance the committed sequence number.
	reload internal.MetaCommitter

	//
	txMgr internal.TxManager

	//============================== external interfaces ==========================================

	// exec is used to execute the block.
	exec external.ExecutionService

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
}

func newOrderRule(oLeader, author uint64, n int, cRecorder internal.CommandRecorder, reader internal.MetaReader, committer internal.MetaCommitter, manager internal.TxManager, exec external.ExecutionService, logger external.Logger) *orderRule {
	return &orderRule{
		author:  author,
		collect: newCollectRule(author, n, cRecorder, logger),
		execute: newExecutionRule(oLeader, author, n, cRecorder, logger),
		commit:  newCommitmentRule(author, n, cRecorder, reader, logger),
		reload:  committer,
		exec:    exec,
		logger:  logger,
		txMgr:   manager,
		commandRecorder: make(map[uint64]uint64),
	}
}

// processPartialOrder is used to process partial order with order rules.
func (rule *orderRule) processPartialOrder() {
	for {
		// order rule 2: execution rule, select commands to execute with natural order.
		frontStream := rule.execute.execution()

		// order rule 3: commitment rule, generate ordered blocks with free will.
		blocks, frontNo := rule.commit.freeWill(frontStream)
		if len(blocks) == 0 {
			// there isn't a committed inner block.
			break
		}

		// commit blocks.
		rule.logger.Debugf("[%d] commit front group, front-no. %d, safe %v, blocks count %d", rule.author, frontNo, frontStream.Safe, len(blocks))
		for _, blk := range blocks {
			rule.seqNo++
			rule.exec.CommandExecution(blk, rule.seqNo)
			rule.reload.Committed(blk.Command.Author, blk.Command.Sequence)
			rule.txMgr.Reply(blk.Command)

			rule.detectFrontSetTypes(!blk.Safe)
			rule.detectFrontAttackGivenRelationship(!blk.Safe, blk.Command)
			rule.detectFrontAttackIntervalRelationship(!blk.Safe, blk.Command)
			rule.updateFrontAttackDetector(blk.Command)
		}
	}
}

func (rule *orderRule) detectFrontSetTypes(risk bool) {
	if !risk {
		rule.totalSafeCommit++
	} else {
		rule.totalRiskCommit++
	}
}

func (rule *orderRule) detectFrontAttackGivenRelationship(risk bool, command *protos.Command) {
	// detect the front attack towards given relationship.
	current := rule.commandRecorder[command.Author]

	if command.Sequence != current+1 {
		if risk {
			rule.frontAttackFromRisk++
		} else {
			rule.frontAttackFromSafe++
		}
	}
}

func (rule *orderRule) detectFrontAttackIntervalRelationship(risk bool, command *protos.Command) {
	// detect the front attack towards interval relationship.
	if command.FrontRunner == nil {
		return
	}

	if command.FrontRunner.Sequence > rule.commandRecorder[command.FrontRunner.Author] {
		if risk {
			rule.frontAttackFromRisk++
			rule.frontAttackIntervalRisk++
		} else {
			rule.frontAttackFromSafe++
			rule.frontAttackIntervalSafe++
		}
	}
}

func (rule *orderRule) updateFrontAttackDetector(command *protos.Command) {
	// update the detector for front attacked command requests.
	if command.Sequence > rule.commandRecorder[command.Author] {
		rule.commandRecorder[command.Author] = command.Sequence
	}
}
