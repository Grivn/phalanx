package executor

import (
	"container/heap"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/executor/types"
	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	author uint64

	last uint64

	pendingLogs *pendingLogs

	cachedLogs *cachedLogs

	executedLogs map[commonTypes.LogID]bool

	recvC chan types.ExecuteLogs

	replyC chan types.ReplyEvent

	closeC chan bool

	logger external.Logger
}

func newExecuteImpl(n int, author uint64, replyC chan types.ReplyEvent, logger external.Logger) *executorImpl {
	return &executorImpl{
		author: author,

		last: uint64(0),

		pendingLogs: newPendingLogs(n, author, logger),

		cachedLogs: newCachedLogs(author, logger),

		executedLogs: make(map[commonTypes.LogID]bool),

		recvC: make(chan types.ExecuteLogs),

		replyC: replyC,

		closeC: make(chan bool),

		logger: logger,
	}
}

func (ei *executorImpl) executeLogs(sequence uint64, logs []*commonProto.OrderedMsg) {
	exec := types.ExecuteLogs{
		Sequence: sequence,
		Logs:     logs,
	}

	ei.recvC <- exec
}

func (ei *executorImpl) start() {
	go ei.listener()
}

func (ei *executorImpl) stop() {
	select {
	case <-ei.closeC:
	default:
		close(ei.closeC)
	}
}

func (ei *executorImpl) listener() {
	for {
		select {
		case <-ei.closeC:
			ei.logger.Notice("exist executor listener")
			return
		case exec := <-ei.recvC:
			ei.processExecuteLogs(exec)
		}
	}
}

func (ei *executorImpl) processExecuteLogs(exec types.ExecuteLogs) {
	ei.cachedLogs.write(exec)

	for {
		orderedLogs := ei.cachedLogs.read()
		if orderedLogs == nil {
			ei.logger.Debugf("replica %d cannot find any ordered-logs on next sequence number", ei.author)
			break
		}

		lh := commonTypes.NewLogHeap()
		for _, orderedLog := range orderedLogs {
			log := ei.pendingLogs.update(orderedLog)
			if log == nil {
				continue
			}
			if ei.executedLogs[log.ID] {
				ei.logger.Debugf("replica %d find an executed log", ei.author)
				continue
			}
			ei.executedLogs[log.ID] = true
			heap.Push(lh, log)
		}

		if lh.Len() > 0 {
			ei.last++
			blk := commonTypes.NewBlock(ei.last, lh.GetSlice())

			for _, log := range blk.Logs {
				ei.pendingLogs.remove(log.ID)
				ei.logger.Debugf("replica %d execute block sequence %d log id %v", ei.author, blk.Sequence, log.ID)
			}

			event := types.ReplyEvent{
				EventType: types.ExecReplyExecuteBlock,
				Event:     blk,
			}
			ei.replyC <- event
		}
	}
}
