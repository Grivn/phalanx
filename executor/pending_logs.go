package executor

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type pendingLogs struct {
	n int

	author uint64

	logs map[commonTypes.LogID]*commonTypes.ExecuteLog

	logger external.Logger
}

func newPendingLogs(n int, author uint64, logger external.Logger) *pendingLogs {
	return &pendingLogs{
		n: n,

		author: author,

		logs: make(map[commonTypes.LogID]*commonTypes.ExecuteLog),

		logger: logger,
	}
}

func (pending *pendingLogs) update(log *commonProto.OrderedLog) {
	id := commonTypes.LogID{
		Author: log.BatchId.Author,
		Hash:   log.BatchId.BatchHash,
	}

	execLog, ok := pending.logs[id]
	if !ok {
		execLog = commonTypes.NewLog(pending.n, id)
		pending.logs[id] = execLog
	}

	execLog.Update(log.Timestamp)
}

func (pending *pendingLogs) remove(id commonTypes.LogID) {
	delete(pending.logs, id)
}
