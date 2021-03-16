package executor

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type pendingLogs struct {
	n int

	author uint64

	logs map[logID]*log

	logger external.Logger
}

func newPendingLogs(n int, author uint64, logger external.Logger) *pendingLogs {
	return &pendingLogs{
		n: n,

		author: author,

		logs: make(map[logID]*log),

		logger: logger,
	}
}

func (pending *pendingLogs) update(msg *commonProto.OrderedMsg) *log {
	id := logID{
		author: msg.BatchId.Author,
		hash:   msg.BatchId.BatchHash,
	}

	l, ok := pending.logs[id]
	if !ok {
		l = newLog(pending.n, id)
		pending.logs[id] = l
	}

	l.update(msg.Timestamp)

	if l.isQuorum() {
		return l
	}

	return nil
}

func (pending *pendingLogs) assign(batch *commonProto.Batch) {
	id := logID{
		author: batch.BatchId.Author,
		hash:   batch.BatchId.BatchHash,
	}

	l, ok := pending.logs[id]
	if !ok {
		l = newLog(pending.n, id)
		pending.logs[id] = l
	}

	l.assign(batch)
}

func (pending *pendingLogs) remove(id logID) {
	delete(pending.logs, id)
}
