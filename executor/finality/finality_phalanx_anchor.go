package finality

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
)

type phalanxAnchorOrdering struct {
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
	reload api.MetaCommitter

	//============================== external interfaces ==========================================

	// exec is used to execute the block.
	exec external.ExecutionService

	// logger is used to print logs.
	logger external.Logger

	// metrics is used to record the metric info of current node's order rule module.
	metrics *metrics.OrderRuleMetrics
}

func newPhalanxAnchorOrdering(conf Config, cRecorder api.CommandRecorder) *phalanxAnchorOrdering {
	return &phalanxAnchorOrdering{
		author:  conf.Author,
		collect: newCollectRule(conf, cRecorder),
		execute: newExecutionRule(conf, cRecorder),
		commit:  newCommitmentRule(conf, cRecorder),
		reload:  conf.Pool,
		exec:    conf.Exec,
		logger:  conf.Logger,
		metrics: conf.Metrics.OrderRuleMetrics,
	}
}

func (pao *phalanxAnchorOrdering) commitOrderStream(oStream types.OrderStream) {
	if len(oStream) == 0 {
		return
	}

	updated := false // if we have updated the command collector.
	for _, oInfo := range oStream {
		// order rule 1: collection rule, collect the partial order info.
		updated = pao.collect.collectPartials(oInfo)
	}

	if updated {
		// if the collector has been updated, try to process the committed partial orders.
		pao.processPartialOrder()
	}
}

// processPartialOrder is used to process partial order with phalanx anchor-based ordering rules.
func (pao *phalanxAnchorOrdering) processPartialOrder() {
	for {
		// order rule 2: execution rule, select commands to execute with natural order.
		frontStream := pao.execute.execution()

		// order rule 3: commitment rule, generate ordered blocks with free will.
		blocks, frontNo := pao.commit.freeWill(frontStream)
		if len(blocks) == 0 {
			// there isn't a committed inner block.
			break
		}

		// commit blocks.
		pao.logger.Debugf("[%d] commit front group, front-no. %d, safe %v, blocks count %d", pao.author, frontNo, frontStream.Safe, len(blocks))
		for _, blk := range blocks {
			pao.seqNo++
			pao.exec.CommandExecution(blk, pao.seqNo)
			pao.reload.Committed(blk.Command.Author, blk.Command.Sequence)

			// record metrics.
			pao.metrics.CommitBlock(blk)
		}
	}
}
