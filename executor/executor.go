package executor

import (
	"sync"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.Mutex

	// seqNo is used to track the sequence number for blocks.
	seqNo uint64

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	// recorder is used to record the command info.
	recorder *commandRecorder

	// exec is used to execute the block.
	exec external.ExecuteService
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(author uint64, n int, exec external.ExecuteService, logger external.Logger) *executorImpl {
	recorder := newCommandRecorder()
	return &executorImpl{
		rules:    newOrderRule(author, n, recorder, logger),
		recorder: recorder,
		exec:     exec,
	}
}

// CommitPartials is used to commit the QCs.
func (ei *executorImpl) CommitPartials(pBatch *protos.PartialOrderBatch) error {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	for _, rawCommand := range pBatch.Commands {
		ei.recorder.storeCommand(rawCommand)
	}

	for _, pOrder := range pBatch.Partials {
		blocks := ei.rules.processPartialOrder(pOrder)
		for _, blk := range blocks {
			ei.seqNo++
			ei.exec.Execute(blk.CommandD, blk.TxList, ei.seqNo, blk.Timestamp)
		}
	}

	return nil
}
