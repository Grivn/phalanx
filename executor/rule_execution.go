package executor

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"

	"github.com/google/btree"
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
	cCommandInfos := er.recorder.readCSCInfos()
	qCommandInfos := er.recorder.readQSCInfos()

	// natural order 1:
	// there isn't any command which has reached correct sequenced status, which means no one could be the pri-command
	// for one command which has reached quorum sequenced status and finished execution for its pri-commands.
	if len(cCommandInfos) == 0 {
		execution = append(execution, qCommandInfos...)
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
		if er.priCheck(qInfo, cCommandInfos) {
			// there isn't any potential priori command.
			execution = append(execution, qInfo)
		}
	}

	return execution
}

func (er *executionRule) priCheck(qInfo *commandInfo, cCommandInfos []*commandInfo) bool {
	valid := true

	// init partial wills map for each participant.
	pWills := make(map[uint64]*btree.BTree)
	for i:=0; i<er.n; i++ {
		pWills[uint64(i+1)] = btree.New(2)
	}

	// put the partial order into it.
	for _, pOrder := range qInfo.pOrders {
		pWills[pOrder.Author()].ReplaceOrInsert(pOrder)
	}

	// pri-command rule:
	//
	// c1: quorum sequenced command
	// c2: correct sequenced command
	// property-priori: c1, c2 belong to replica's collected partial order list
	//                  and c2<-c1
	//
	// if the amount of replica with property-priori is no less than f+1, we regard c2 as c1's pri-command.
	for _, cInfo := range cCommandInfos {
		for _, pOrder := range cInfo.pOrders {
			pWills[pOrder.Author()].ReplaceOrInsert(pOrder)
		}

		count := 0
		for _, pWill := range pWills {
			item := pWill.Min()

			if item == nil {
				continue
			}

			pOrder := item.(*protos.PartialOrder)

			if pOrder.CommandDigest() == cInfo.curCmd {
				count++
			}
		}

		if count >= er.oneCorrect {
			valid = false
			qInfo.prioriRecord(cInfo.curCmd)
			er.logger.Debugf("[%d] potential natural order: %s <- %s", er.author, cInfo.format(), qInfo.format())
		}

		for _, pOrder := range cInfo.pOrders {
			pWills[pOrder.Author()].Delete(pOrder)
		}
	}

	// we have selected all the potential priori commands.
	qInfo.trusted = true

	if !valid {
		er.recorder.potentialByz(qInfo)
	}

	return valid
}
