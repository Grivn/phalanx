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

func (er *executionRule) naturalOrder() []*CommandInfo {
	// here, we would like to check the natural order for quorum sequenced commands.
	var execution []*CommandInfo

	qCommandInfos := er.recorder.ReadQSCInfos()

	cCommandInfos := er.recorder.ReadCSCInfos()

	wCommandInfos := er.recorder.ReadWatInfos()

	// natural order 1:
	// there isn't any command which has reached correct sequenced status, which means no one could be the pri-command
	// for one command which has reached quorum sequenced status and finished execution for its pri-commands.
	if len(cCommandInfos) == 0 && len(wCommandInfos) == 0 {
		execution = append(execution, qCommandInfos...)
		return execution
	}

	// natural order 2&3:
	for _, qInfo := range qCommandInfos {
		if qInfo.Trust {
			// we have selected all the potential priori commands.
			execution = append(execution, qInfo)
			continue
		}

		// the potential priority check between quorum sequenced command and waiting command
		// here, there is possibility for us to generate Condorcet Paradox, so that we should resolve the cyclic problems.
		//
		// the potential priority check between quorum sequenced command and correct sequenced command
		// here, we only need to verify the priority properties for these correct sequenced command do not have any priority.
		if er.priorityCheck(qInfo, wCommandInfos, cCommandInfos) {
			// there isn't any potential priori command.
			execution = append(execution, qInfo)
		}
	}

	return execution
}

func (er *executionRule) priorityCheck(qInfo *CommandInfo, wInfos []*CommandInfo, cInfos []*CommandInfo) bool {
	defer func(qInfo *CommandInfo) {
		// current command cannot be a leaf node,
		// for which it should be selected into execution list or have some priorities.
		er.recorder.CutLeaf(qInfo)
	}(qInfo)

	var newPriorities []string

	qPointers := make(map[uint64]uint64)

	// initiate the pointer for quorum replicas.
	for _, pOrder := range qInfo.Orders {
		qPointers[pOrder.Author()] = pOrder.Sequence()
	}
	for _, wInfo := range wInfos {
		if qInfo.PriCmd[wInfo.CurCmd] {
			// it has already become the priority command of QSC.
			continue
		}

		count := 0

		// check if there are f+1 replicas believe current QSC should be selected before waiting command.
		for id, seq := range qPointers {
			pOrder, ok := wInfo.Orders[id]
			if !ok || pOrder.Sequence() < seq {
				count++
			}
			if count == er.oneCorrect {
				break
			}
		}

		if count < er.oneCorrect {
			// should make sure a Condorcet Paradox wouldn't occur.
			helper := NewScanner(qInfo)

			// only the command in leaf nodes could be involved into cyclic dependency.
			if er.recorder.IsLeaf(qInfo) {
				// if current waiting command has a leaf node equal to current one, a cyclic dependency occurs.
				if helper.Scan() {
					er.logger.Debugf("[%d] priority command depend on self %s", er.author, qInfo.Format())
					continue
				}
			}

			newPriorities = append(newPriorities, wInfo.CurCmd)
			er.logger.Debugf("[%d] potential natural order: %s <- %s", er.author, wInfo.Format(), qInfo.Format())
		}
	}

	for _, cInfo := range cInfos {
		if qInfo.PriCmd[cInfo.CurCmd] {
			// it has already become the priority command of QSC.
			continue
		}

		count := 0

		for id, seq := range qPointers {
			pOrder, ok := cInfo.Orders[id]
			if !ok || pOrder.Sequence() < seq {
				count++
			}
			if count == er.oneCorrect {
				break
			}
		}

		if count < er.oneCorrect {
			// record the priority command.
			qInfo.PrioriRecord(cInfo)

			// the priority command should become a leaf node,
			// for which it does not have any prefix commands and has become other command's priority.
			er.recorder.AddLeaf(cInfo)

			// update the priority list.
			newPriorities = append(newPriorities, cInfo.CurCmd)
			er.logger.Debugf("[%d] potential natural order: %s <- %s", er.author, cInfo.Format(), qInfo.Format())

			qInfo.AppendLow(cInfo)
		}
	}

	if len(qInfo.PriCmd) > 0 {
		er.recorder.PotentialByz(qInfo, newPriorities)
		return false
	}

	return true
}
