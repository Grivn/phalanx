package executor

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"sort"
)

type executorImpl struct {
	n int

	author uint64

	last uint64

	pendingLogs *pendingLogs

	cachedLogs *cachedLogs

	executedLogs map[commonTypes.LogID]bool

	recvC commonTypes.ExecutorRecvChan

	sendC commonTypes.ExecutorSendChan

	closeC chan bool

	logger external.Logger
}

func newExecuteImpl(n int, author uint64, sendC commonTypes.ExecutorSendChan, logger external.Logger) *executorImpl {
	recvC := commonTypes.ExecutorRecvChan{
		ExecuteLogsChan: make(chan *commonProto.ExecuteLogs),
	}

	return &executorImpl{
		n: n,

		author: author,

		last: uint64(0),

		pendingLogs: newPendingLogs(n, author, logger),

		cachedLogs: newCachedLogs(author, logger),

		executedLogs: make(map[commonTypes.LogID]bool),

		recvC: recvC,

		sendC: sendC,

		closeC: make(chan bool),

		logger: logger,
	}
}

func (ei *executorImpl) executeLogs(exec *commonProto.ExecuteLogs) {
	ei.recvC.ExecuteLogsChan <- exec
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
		case exec := <-ei.recvC.ExecuteLogsChan:
			ei.processExecuteLogs(exec)
		}
	}
}

func (ei *executorImpl) processExecuteLogs(exec *commonProto.ExecuteLogs) {
	ei.cachedLogs.write(exec)

	for {
		orderedLogs := ei.cachedLogs.read()
		if orderedLogs == nil {
			ei.logger.Debugf("replica %d cannot find any ordered-logs on next sequence number", ei.author)
			break
		}

		for _, orderedLog := range orderedLogs {
			if ei.executedLogs[commonTypes.LogID{Author: orderedLog.BatchId.Author, Hash:orderedLog.BatchId.BatchHash}] {
				ei.logger.Debugf("replica %d find an executed log", ei.author)
				continue
			}
			ei.pendingLogs.update(orderedLog)
		}

		var eSlice []*commonTypes.ExecuteLog
		for id, executeLog := range ei.pendingLogs.logs {
			ei.logger.Infof("replica %d check the log, %v, len %d", ei.author, id, executeLog.Len())
			if executeLog.IsQuorum() {
				eSlice = append(eSlice, executeLog)
				ei.pendingLogs.remove(id)
				ei.executedLogs[id] = true
			}
		}

		if len(eSlice) > 0 {
			ei.last++
			blk := commonTypes.NewBlock(ei.last, eSlice)
			sort.Sort(blk)
			blk.UpdateTimestamp()

			for _, log := range blk.Logs {
				ei.logger.Debugf("replica %d execute block sequence %d log id %v", ei.author, blk.Sequence, log.ID)
			}
			ei.sendC.BlockChan <- blk
		}
	}
}
