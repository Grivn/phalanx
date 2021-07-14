package executor

import (
	"github.com/Grivn/phalanx/internal"
	"sync"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.Mutex

	// author indicates the identifier of current node.
	author uint64

	// seqNo is used to track the sequence number for blocks.
	seqNo uint64

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	// recorder is used to record the command info.
	recorder *commandRecorder

	//
	mgr internal.LogManager

	// exec is used to execute the block.
	exec external.ExecutionService
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(author uint64, n int, mgr internal.LogManager, exec external.ExecutionService, logger external.Logger) *executorImpl {
	recorder := newCommandRecorder(author, logger)
	return &executorImpl{
		author:   author,
		rules:    newOrderRule(author, n, recorder, logger),
		recorder: recorder,
		exec:     exec,
		mgr:      mgr,
	}
}

// CommitPartials is used to commit the QCs.
func (ei *executorImpl) CommitPartials(pBatch *protos.PartialOrderBatch) error {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	if pBatch == nil {
		// nil partial order batch means we should skip the current commitment attempt.
		return nil
	}

	for _, rawCommand := range pBatch.Commands {
		ei.recorder.storeCommand(rawCommand)
	}

	for _, pOrder := range pBatch.Partials {
		blocks := ei.rules.processPartialOrder(pOrder)
		for _, blk := range blocks {
			ei.seqNo++
			ei.exec.CommandExecution(blk.CommandD, blk.TxList, ei.seqNo, blk.Timestamp)
			ei.mgr.Committed(blk.Author, blk.CmdSeq)
		}
	}

	return nil
}
