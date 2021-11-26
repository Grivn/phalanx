package execsimple

import (
	"github.com/Grivn/phalanx/internal"
	"sort"

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

func (collect *collectionRule) collectPartials(oInfo types.OrderInfo) bool {
	collect.logger.Infof("[%d] collect partial order: %s", collect.author, oInfo.Format())

	// find the digest for current command the partial order refers to.
	commandD := oInfo.Command

	// check if current command has been committed or not.
	if collect.cRecorder.IsCommitted(commandD) {
		collect.logger.Debugf("[%d] committed command %s, ignore it", collect.author, commandD)
		return false
	}

	// push back partial order into recorder.queue.
	if err := collect.cRecorder.PushBack(oInfo); err != nil {
		collect.logger.Errorf("[%d] push back partial order failed: %s", collect.author, err)
		return false
	}

	// read command info from command cRecorder.
	info := collect.cRecorder.ReadCommandInfo(commandD)
	info.OrderAppend(oInfo)

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
