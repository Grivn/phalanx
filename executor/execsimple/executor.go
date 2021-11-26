package execsimple

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/recorder"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"sort"
	"sync"
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

	//
	orderSeq map[uint64]uint64

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

func NewExecutor(author uint64, n int, mgr internal.MetaPool, exec external.ExecutionService, logger external.Logger) *executorImpl {
	orderSeq := make(map[uint64]uint64)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		orderSeq[id] = uint64(0)
	}

	cRecorder := recorder.NewCommandRecorder(author, n, logger)
	return &executorImpl{
		author:    author,
		rules:     newOrderRule(author, n, cRecorder, mgr, logger),
		cRecorder: cRecorder,
		exec:      exec,
		committer: mgr,
		reader:    mgr,
		logger:    logger,
		orderSeq:  orderSeq,
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

	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	partials := ei.reader.ReadPartials(qStream)

	var oStream types.OrderStream

	for _, pOrder := range partials {
		startNo := ei.orderSeq[pOrder.Author()]

		infos, endNo := types.NewOrderInfos(startNo, pOrder)

		ei.orderSeq[pOrder.Author()] = endNo
		oStream = append(oStream, infos...)
	}
	sort.Sort(oStream)
	ei.logger.Debugf("[%d] commit order info stream len %d: %v", ei.author, len(oStream), oStream)

	for _, oInfo := range oStream {
		blocks := ei.rules.processPartialOrder(oInfo)
		for _, blk := range blocks {
			ei.seqNo++
			ei.exec.CommandExecution(blk.Command, ei.seqNo, blk.Timestamp)
			ei.committer.Committed(blk.Command.Author, blk.Command.Sequence)
		}
	}

	return nil
}
