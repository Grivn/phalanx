package finality

import (
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/metrics"
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

	//
	metrics *metrics.OrderRuleMetrics

	//
	mediumCommit *orderMediumT
}

func newOrderRule(conf Config, cRecorder internal.CommandRecorder) *orderRule {
	return &orderRule{
		author:       conf.Author,
		collect:      newCollectRule(conf, cRecorder),
		execute:      newExecutionRule(conf, cRecorder),
		commit:       newCommitmentRule(conf, cRecorder),
		reload:       conf.Mgr,
		exec:         conf.Exec,
		logger:       conf.Logger,
		txMgr:        conf.Manager,
		mediumCommit: newOrderMediumT(conf),
		metrics:      conf.Metrics.OrderRuleMetrics,
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

			// commit blocks with medium timestamp.
			rule.mediumCommit.commitAccordingMediumT(blk)

			// record metrics.
			rule.metrics.CommitBlock(blk)
		}
	}
}
