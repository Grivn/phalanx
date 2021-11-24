package execcomplex

import (
	"sort"
	"sync"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/recorder"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
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

	// cRecorder is used to record the command info.
	cRecorder internal.CommandRecorder

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	//============================= internal interfaces =========================================

	// committer is used to notify client instance the committed sequence number.
	committer internal.MetaCommitter

	// reader is used to read partial orders from meta pool tracker.
	reader internal.MetaReader

	//============================== external interfaces ==========================================

	// exec is used to execute the block.
	exec external.ExecutionService

	// logger is used to print logs.
	logger external.Logger
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(author uint64, n int, mgr internal.MetaPool, exec external.ExecutionService, logger external.Logger) internal.Executor {
	cRecorder := recorder.NewCommandRecorder(author, n, logger)
	return &executorImpl{
		author:    author,
		rules:     newOrderRule(author, n, cRecorder, mgr, logger),
		cRecorder: cRecorder,
		exec:      exec,
		committer: mgr,
		reader:    mgr,
		logger:    logger,
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
		blocks := ei.rules.processPartialOrder(pOrder)
		for _, blk := range blocks {
			ei.seqNo++
			ei.exec.CommandExecution(blk.Command, ei.seqNo, blk.Timestamp)
			ei.committer.Committed(blk.Command.Author, blk.Command.Sequence)
		}
	}

	return nil
}
