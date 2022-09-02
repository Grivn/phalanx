package finality

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/barrier"
	"github.com/Grivn/phalanx/external"
)

type executionRule struct {
	// preTag
	preTag bool

	// author indicates the identifier of current node.
	author uint64

	// n indicates the number of replicas.
	n int

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// cRecorder is used to record the command info.
	cRecorder api.CommandRecorder

	selected map[string]bool

	// logger is used to print logs.
	logger external.Logger

	// oligarchy is used to define that current cluster is relying on a certain node.
	oligarchy uint64
}

func newExecutionRule(conf Config, recorder api.CommandRecorder) *executionRule {
	conf.Logger.Infof("[%d] initiate natural order handler, replica count %d", conf.Author, conf.N)
	return &executionRule{
		preTag:     true,
		author:     conf.Author,
		n:          conf.N,
		oneCorrect: types.CalculateOneCorrect(conf.N),
		quorum:     types.CalculateQuorum(conf.N),
		cRecorder:  recorder,
		logger:     conf.Logger,
		oligarchy:  conf.OLeader,
	}
}

func (er *executionRule) execution() types.FrontStream {

	// oligarchy mode, relying on certain leader ordering.
	if er.oligarchy != uint64(0) {
		return er.oligarchyExecution()
	}

	// read the front set.
	commands, safe := er.cRecorder.FrontCommands()

	var cStream types.CommandStream
	for _, digest := range commands {
		info := er.cRecorder.ReadCommandInfo(digest)
		cStream = append(cStream, info)
	}

	if !safe {
		// we cannot make sure the validation of front set.
		handler := barrier.NewCommandStreamBarrier(er.author, er.cRecorder, er.oneCorrect, er.logger)
		cStream = handler.BaselineGroup(cStream)
	}

	return types.FrontStream{Safe: safe, Stream: cStream}
}

func (er *executionRule) oligarchyExecution() types.FrontStream {
	digest := er.cRecorder.OligarchyLeaderFront(er.oligarchy)
	commandInfo := er.cRecorder.ReadCommandInfo(digest)
	if len(commandInfo.Orders) < er.quorum {
		return types.FrontStream{Safe: true, Stream: nil}
	}
	return types.FrontStream{Safe: true, Stream: types.CommandStream{commandInfo}}
}
