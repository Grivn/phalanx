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
	return &executionRule{
		author:     author,
		n:          n,
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		recorder:	recorder,
		logger:     logger,
	}
}

func (er *executionRule) executeQSCs() []*commandInfo {
	var execution []*commandInfo
	cCommandInfos := er.recorder.readCSCInfos()
	qCommandInfos := er.recorder.readQSCInfos()

	// case 1:
	// there isn't any command which has reached correct sequenced status, here, we could execute
	if len(cCommandInfos) == 0 {
		execution = append(execution, qCommandInfos...)
		return execution
	}

	// case 2:
	// check the validation for every quorum sequenced command that make sure all the commands in
	// correct sequenced status cannot become the pri-command of it.
	for _, qInfo := range qCommandInfos {
		if er.priCheck(qInfo, cCommandInfos) {
			execution = append(execution, qInfo)
		}
	}

	return execution
}

func (er *executionRule) priCheck(qInfo *commandInfo, cCommandInfos []*commandInfo) bool {
	valid := true

	// init filter btree map
	filter := make(map[uint64]*btree.BTree)
	for i:=0; i<er.n; i++ {
		filter[uint64(i+1)] = btree.New(2)
	}

	// put the partial order into it.
	for _, pOrder := range qInfo.pOrders {
		filter[pOrder.Author()].ReplaceOrInsert(pOrder)
	}

	for _, cInfo := range cCommandInfos {
		for _, pOrder := range cInfo.pOrders {
			filter[pOrder.Author()].ReplaceOrInsert(pOrder)
		}

		count := 0
		for _, will := range filter {
			item := will.Min()

			if item == nil {
				continue
			}

			orderWill := item.(*protos.PartialOrder)

			if orderWill.Digest() == cInfo.curCmd {
				count++
			}
		}

		if count >= er.oneCorrect {
			valid = false
			qInfo.prioriRecord(cInfo.curCmd)
		}

		for _, pOrder := range cInfo.pOrders {
			filter[pOrder.Author()].Delete(pOrder)
		}
	}

	return valid
}
