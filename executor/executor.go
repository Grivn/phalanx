package executor

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/internal"
	"sort"
	"sync"

	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.Mutex

	//============================ basic information =============================================

	// author indicates the identifier of current node.
	author uint64

	// seqNo is used to track the sequence number for blocks.
	seqNo uint64

	//============================ order rule for block generation ========================================

	// commandMap is used to record the hash of commands which have been selected into executor.
	commandMap map[string]bool

	// recorder is used to record the command info.
	recorder *commandRecorder

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	//============================= internal interfaces =========================================

	// committer is used to notify client instance the committed sequence number.
	committer internal.Committer

	// reader is used to read commands and partial orders from the tracker.
	reader internal.Reader

	//============================== external interfaces ==========================================

	// exec is used to execute the block.
	exec external.ExecutionService

	// logger is used to print logs.
	logger external.Logger
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(author uint64, n int, mgr internal.MetaPool, exec external.ExecutionService, logger external.Logger) *executorImpl {
	recorder := newCommandRecorder(author, logger)
	return &executorImpl{
		author:     author,
		rules:      newOrderRule(author, n, recorder, logger),
		recorder:   recorder,
		exec:       exec,
		committer:  mgr,
		reader:     mgr,
		commandMap: make(map[string]bool),
		logger:     logger,
	}
}

// CommitStream is used to commit the partial order stream.
func (ei *executorImpl) CommitStream(qStream types.QueryStream) error {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	if len(qStream) == 0 {
		// nil partial order batch means we should skip the current commitment attempt.
		return nil
	}

	sort.Sort(qStream)
	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	partials := ei.reader.ReadPartials(qStream)

	for _, pOrder := range partials {

		if !ei.commandMap[pOrder.CommandDigest()] {
			// raed the command info.
			command := ei.reader.ReadCommand(pOrder.CommandDigest())
			ei.recorder.storeCommand(command)
			ei.commandMap[pOrder.CommandDigest()] = true
		}

		blocks := ei.rules.processPartialOrder(pOrder)
		for _, blk := range blocks {
			ei.seqNo++
			ei.exec.CommandExecution(blk.Command, ei.seqNo, blk.Timestamp)
			ei.committer.Committed(blk.Command.Author, blk.Command.Sequence)
		}
	}

	return nil
}
