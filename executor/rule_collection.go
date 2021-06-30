package executor

import (
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

	// recorder is used to record the command info.
	recorder *commandRecorder

	// logger is used to print logs.
	logger external.Logger
}

func newCollectRule(author uint64, n int, recorder *commandRecorder, logger external.Logger) *collectionRule {
	return &collectionRule{
		author:     author,
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		recorder:	recorder,
		logger:     logger,
	}
}

func (collect *collectionRule) collectPartials(pOrder *protos.PartialOrder) {
	collect.logger.Infof("replica %d collect partial order for replica %d sequence %d digest %s", collect.author, pOrder.Author(), pOrder.Sequence(), pOrder.CommandDigest())

	// find the digest for current command the partial order refers to.
	commandD := pOrder.CommandDigest()

	// check if current command has been committed or not.
	if collect.recorder.isCommitted(commandD) {
		collect.logger.Debugf("replica %d has already committed command %s, ignore it", collect.author, commandD)
		return
	}

	// read command info from command recorder.
	info := collect.recorder.readCommandInfo(commandD)

	if info.pOrderCount() >= collect.quorum {
		// for one command, we only need to collect the partial orders from quorum replicas, ignore the redundant partial order.
		collect.logger.Infof("replica %d has collected partial orders from quorum replicas for command %s, ignore it", collect.author, commandD)
		return
	}
	info.pOrderAppend(pOrder)

	// check the command status.
	switch info.pOrderCount() {
	case collect.oneCorrect:
		// current command has reached correct sequenced status.
		collect.recorder.correctStatus(commandD)
		collect.logger.Infof("replica %d has collected at least one correct partial order for command %s", collect.author, commandD)
	case collect.quorum:
		// current command has reached quorum sequenced status.
		sort.Sort(info.timestamps)
		collect.recorder.quorumStatus(commandD)
		collect.logger.Infof("replica %d has collected quorum partial order for command %s", collect.author, commandD)
	}
}
