package executor

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type orderRule struct {
	collect *collectionRule
	execute *executionRule
	commit  *commitmentRule
}

func newOrderRule(author uint64, n int, recorder *commandRecorder, logger external.Logger) *orderRule {
	return &orderRule{
		collect: newCollectRule(author, n, recorder, logger),
		execute: newExecutionRule(author, n, recorder, logger),
		commit:  newCommitmentRule(author, n, recorder, logger),
	}
}

func (rule *orderRule) processPartialOrder(pOrder *protos.PartialOrder) []types.Block {
	// order rule 1: collection rule.
	rule.collect.collectPartials(pOrder)

	// order rule 2: execution rule.
	executionList := rule.execute.executeQSCs()

	// order rule 3: commitment rule.
	blocks := rule.commit.freeWill(executionList)

	return blocks
}
