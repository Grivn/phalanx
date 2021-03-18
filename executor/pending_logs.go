package executor

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type pendingLogs struct {
	n int

	author uint64

	logs map[commonTypes.LogID]*commonTypes.Log

	logger external.Logger
}

func newPendingLogs(n int, author uint64, logger external.Logger) *pendingLogs {
	return &pendingLogs{
		n: n,

		author: author,

		logs: make(map[commonTypes.LogID]*commonTypes.Log),

		logger: logger,
	}
}

func (pending *pendingLogs) update(msg *commonProto.OrderedMsg) *commonTypes.Log {
	id := commonTypes.LogID{
		Author: msg.BatchId.Author,
		Hash:   msg.BatchId.BatchHash,
	}

	log, ok := pending.logs[id]
	if !ok {
		log = commonTypes.NewLog(pending.n, id)
		pending.logs[id] = log
	}

	log.Update(msg.Timestamp)

	if log.IsQuorum() {
		return log
	}

	return nil
}

func (pending *pendingLogs) remove(id commonTypes.LogID) {
	delete(pending.logs, id)
}
