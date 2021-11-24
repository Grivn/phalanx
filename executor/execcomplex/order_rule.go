package execcomplex

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

// orderRule contains the essential rules for block generation.
type orderRule struct {
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
}

func newOrderRule(author uint64, n int, cRecorder internal.CommandRecorder, reader internal.MetaReader, logger external.Logger) *orderRule {
	return &orderRule{
		collect: newCollectRule(author, n, cRecorder, logger),
		execute: newExecutionRule(author, n, cRecorder, logger),
		commit:  newCommitmentRule(author, n, cRecorder, reader, logger),
	}
}

// processPartialOrder is used to process partial order with order rules.
func (rule *orderRule) processPartialOrder(pOrder *protos.PartialOrder) []types.InnerBlock {
	// order rule 1: collection rule, collect the partial order.
	if collected := rule.collect.collectPartials(pOrder); !collected {
		return nil
	}

	var blocks []types.InnerBlock
	for {
		// order rule 2: execution rule, select commands to execute with natural order.
		executionList := rule.execute.naturalOrder()

		if len(executionList) == 0 {
			// there isn't an executable sequenced command.
			break
		}

		// order rule 3: commitment rule, generate ordered blocks with free will.
		blocks = append(blocks, rule.commit.freeWill(executionList)...)
	}


	return blocks
}
