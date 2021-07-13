package executor

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type executionRule struct {
	// author indicates the identifier of current node.
	author uint64

	// n indicates the number of replicas.
	n int

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// recorder is used to record the command info.
	recorder *commandRecorder

	// logger is used to print logs.
	logger external.Logger
}

func newExecutionRule(author uint64, n int, recorder *commandRecorder, logger external.Logger) *executionRule {
	logger.Infof("[%d] initiate natural order handler, replica count %d", author, n)
	return &executionRule{
		author:     author,
		n:          n,
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		recorder:	recorder,
		logger:     logger,
	}
}

func (er *executionRule) naturalOrder() []*commandInfo {
	// here, we would like to check the natural order for quorum sequenced commands.
	var execution []*commandInfo

	qCommandInfos := er.recorder.readQSCInfos()
	potentialInfos := append(er.recorder.readCSCInfos(), er.recorder.readWatInfos()...)

	// natural order 1:
	// there isn't any command which has reached correct sequenced status, which means no one could be the pri-command
	// for one command which has reached quorum sequenced status and finished execution for its pri-commands.
	if len(potentialInfos) == 0 {
		for _, qInfo := range qCommandInfos {
			execution = append(execution, qInfo)
		}
		return execution
	}

	// natural order 2:
	// check the quorum sequenced command to make sure all the commands in correct sequenced status cannot become
	// the pri-command of it.
	for _, qInfo := range qCommandInfos {
		if qInfo.trusted {
			// we have selected all the potential priori commands.
			execution = append(execution, qInfo)
			continue
		}
		if er.priorityCheck(qInfo, potentialInfos) {
			// there isn't any potential priori command.
			execution = append(execution, qInfo)
		}
	}

	return execution
}

func (er *executionRule) priorityCheck(qInfo *commandInfo, checkInfos []*commandInfo) bool {
	qPointers := make(map[uint64]uint64)

	// initiate the pointer for quorum replicas.
	for _, pOrder := range qInfo.pOrders {
		qPointers[pOrder.Author()] = pOrder.Sequence()
	}

	for _, checkInfo := range checkInfos {
		count := 0

		for id, seq := range qPointers {
			pOrder, ok := checkInfo.pOrders[id]
			if !ok || pOrder.Sequence() < seq {
				count++
			}
			if count == er.oneCorrect {
				break
			}
		}

		if count < er.oneCorrect {
			if er.recorder.isLeaf(qInfo.curCmd) {
				er.logger.Debugf("[%d] priority command depend on self %s", er.author, qInfo.format())
				continue
			}

			qInfo.prioriRecord(checkInfo)
			er.logger.Debugf("[%d] potential natural order: %s <- %s", er.author, checkInfo.format(), qInfo.format())

			if checkInfo.pOrderCount() < er.quorum {
				er.recorder.recordLeaf(checkInfo.curCmd)
			}
		}
	}
	// we have selected all the potential priori commands.
	qInfo.trusted = true

	if len(qInfo.priCmd) > 0 {
		er.recorder.potentialByz(qInfo)
		er.recorder.removeLeaf(qInfo.curCmd)
		return false
	}

	return true
}
