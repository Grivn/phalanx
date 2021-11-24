package execcomplex

import (
	"github.com/Grivn/phalanx/internal"
	"sort"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type collectionRule struct {
	// author indicates the identifier of current node.
	author uint64

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// cRecorder is used to record the command info.
	cRecorder internal.CommandRecorder

	// logger is used to print logs.
	logger external.Logger
}

func newCollectRule(author uint64, n int, recorder internal.CommandRecorder, logger external.Logger) *collectionRule {
	logger.Infof("[%d] initiate partial order collector, replica count %d", author, n)
	return &collectionRule{
		author:     author,
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		cRecorder:  recorder,
		logger:     logger,
	}
}

func (collect *collectionRule) collectPartials(pOrder *protos.PartialOrder) bool {
	collect.logger.Infof("[%d] collect partial order: %s", collect.author, pOrder.Format())

	// find the digest for current command the partial order refers to.
	commandD := pOrder.CommandDigest()

	// check if current command has been committed or not.
	if collect.cRecorder.IsCommitted(commandD) {
		collect.logger.Debugf("[%d] committed command %s, ignore it", collect.author, commandD)
		return false
	}

	// read command info from command cRecorder.
	info := collect.cRecorder.ReadCommandInfo(commandD)

	if info.OrderCount() >= collect.quorum {
		// for one command, we only need to collect the partial orders from quorum replicas, ignore the redundant partial order.
		collect.logger.Debugf("[%d] command %s in quorum sequenced status, ignore it", collect.author, commandD)
		return false
	}
	info.OrderAppend(pOrder)

	// check the command status.
	switch info.OrderCount() {
	case collect.oneCorrect:
		// current command has reached correct sequenced status.
		collect.cRecorder.CorrectStatus(commandD)
		collect.logger.Infof("[%d] found correct sequenced command %s", collect.author, commandD)
	case collect.quorum:
		// current command has reached quorum sequenced status.
		sort.Sort(info.Timestamps)
		collect.cRecorder.QuorumStatus(commandD)
		collect.logger.Infof("[%d] found quorum sequenced command %s", collect.author, commandD)
	}
	return true
}
