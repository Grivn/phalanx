package executor

import (
	"container/list"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/executor/types"
	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	author uint64

	last uint64

	blockList *list.List

	pendingLogs *pendingLogs

	cachedLogs *cachedLogs

	recvC chan types.RecvEvent

	replyC chan types.ReplyEvent

	closeC chan bool

	executor external.Executor

	logger external.Logger
}

func newExecuteImpl(n int, author uint64, replyC chan types.ReplyEvent, executor external.Executor, logger external.Logger) *executorImpl {
	return &executorImpl{
		author: author,

		last: uint64(0),

		blockList: list.New(),

		pendingLogs: newPendingLogs(n, author, logger),

		cachedLogs: newCachedLogs(author, logger),

		recvC: make(chan types.RecvEvent),

		replyC: replyC,

		closeC: make(chan bool),

		executor: executor,

		logger: logger,
	}
}

func (ei *executorImpl) executeLogs(sequence uint64, logs []*commonProto.OrderedMsg) {
	exec := types.ExecuteLogs{
		Sequence: sequence,
		Logs:     logs,
	}
	event := types.RecvEvent{
		EventType: types.ExecRecvLogs,
		Event:     exec,
	}

	ei.recvC <- event
}

func (ei *executorImpl) executeBatch(batch *commonProto.Batch) {
	event := types.RecvEvent{
		EventType: types.ExecRecvBatch,
		Event:     batch,
	}

	ei.recvC <- event
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
		case ev := <-ei.recvC:
			ei.dispatchEvent(ev)
		}
	}
}

func (ei *executorImpl) dispatchEvent(event types.RecvEvent) {
	switch event.EventType {
	case types.ExecRecvLogs:
		exec := event.Event.(types.ExecuteLogs)
		ei.cachedLogs.write(exec)

		for {
			orderedLogs := ei.cachedLogs.read()
			if orderedLogs == nil {
				break
			}

			finishedLogs := make(map[logID]*log)
			for _, orderedLog := range orderedLogs {
				l := ei.pendingLogs.update(orderedLog)
				if l == nil {
					continue
				}
				finishedLogs[l.id] = l
			}

			if len(finishedLogs) > 0 {
				var slice []*log
				for _, finishedLog := range finishedLogs {
					slice = append(slice, finishedLog)
				}
				blk := newBlock(slice)
				e := ei.blockList.PushBack(blk)

				if ei.tryToExecute(blk) {
					ei.blockList.Remove(e)
				}
			}
		}
	case types.ExecRecvBatch:
		batch := event.Event.(*commonProto.Batch)
		ei.pendingLogs.assign(batch)

		for {
			if ei.blockList.Len() == 0 {
				break
			}
			e := ei.blockList.Front()
			blk := e.Value.(*block)

			if ei.tryToExecute(blk) {
				ei.blockList.Remove(e)
			}
		}
	default:
		ei.logger.Errorf("Invalid event type: code %d", event.EventType)
		return
	}
}

func (ei *executorImpl) tryToExecute(blk *block) bool {
	for _, l := range blk.logs {
		if !l.assigned() {
			bid := &commonProto.BatchId{
				Author:    l.id.author,
				BatchHash: l.id.hash,
			}
			event := types.ReplyEvent{
				EventType: types.ExecReplyLoadBatch,
				Event:     bid,
			}
			ei.replyC <- event
			return false
		}
	}

	var txs []*commonProto.Transaction
	var localList []bool
	var local bool

	for _, l := range blk.logs {
		if l.batch.BatchId.Author == ei.author {
			local = true
		} else {
			local = false
		}

		for _, tx := range l.batch.TxList {
			txs = append(txs, tx)
			localList = append(localList, local)
		}
	}

	ei.last++
	ei.logger.Noticef("======== replica %d call execute, seqNo=%d", ei.author, ei.last)
	ei.executor.Execute(txs, localList, ei.last, blk.timestamp)

	for _, l := range blk.logs {
		ei.pendingLogs.remove(l.id)
	}

	return true
}
