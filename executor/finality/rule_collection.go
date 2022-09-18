package finality

import (
	"github.com/Grivn/phalanx/common/api"
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
	cRecorder api.CommandRecorder

	// logger is used to print logs.
	logger external.Logger
}

func newCollectRule(conf Config, recorder api.CommandRecorder) *collectionRule {
	conf.Logger.Infof("[%d] initiate partial order collector, replica count %d", conf.Author, conf.N)
	return &collectionRule{
		author:     conf.Author,
		oneCorrect: types.CalculateOneCorrect(conf.N),
		quorum:     types.CalculateQuorum(conf.N),
		cRecorder:  recorder,
		logger:     conf.Logger,
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

	// already committed by quorum replicas, then update the timestamp list.
	if collect.cRecorder.IsQuorum(commandD) {
		info.UpdateTrustedTS(collect.oneCorrect)
	}

	// check the command status.
	switch info.OrderCount() {
	case collect.oneCorrect:
		// current command has reached correct sequenced status.
		collect.cRecorder.CorrectStatus(commandD)
		collect.logger.Infof("[%d] found correct sequenced command %s", collect.author, commandD)
	case collect.quorum:
		// current command has reached quorum sequenced status.
		collect.cRecorder.QuorumStatus(commandD)
		collect.logger.Infof("[%d] found quorum sequenced command %s", collect.author, commandD)
		info.UpdateTrustedTS(collect.oneCorrect)
		info.UpdateMediumTS(collect.oneCorrect)
	}
	return true
}
