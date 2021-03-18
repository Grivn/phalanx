package executor

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/executor/types"
	"github.com/Grivn/phalanx/external"
)

type cachedLogs struct {
	author uint64

	lastSeq uint64

	logs map[uint64][]*commonProto.OrderedMsg

	logger external.Logger
}

func newCachedLogs(author uint64, logger external.Logger) *cachedLogs {
	return &cachedLogs{
		author: author,

		lastSeq: uint64(0),

		logs: make(map[uint64][]*commonProto.OrderedMsg),

		logger: logger,
	}
}

func (cached *cachedLogs) write(exec types.ExecuteLogs) {
	if exec.Sequence <= cached.lastSeq {
		cached.logger.Warningf("Invalid sequence number, now last sequence %d", cached.lastSeq)
		return
	}

	cached.logger.Debugf("replica %d cached the logs of sequence %d", cached.author, exec.Sequence)
	cached.logs[exec.Sequence] = exec.Logs
}

func (cached *cachedLogs) read() []*commonProto.OrderedMsg {
	sequence := cached.lastSeq+1
	cached.logger.Debugf("replica %d try to read the logs of sequence %d", cached.author, sequence)

	logs, ok := cached.logs[sequence]
	if !ok {
		cached.logger.Debugf("replica %d cannot find the logs on sequence %d", cached.author, sequence)
		return nil
	}

	delete(cached.logs, sequence)
	cached.lastSeq++
	return logs
}
