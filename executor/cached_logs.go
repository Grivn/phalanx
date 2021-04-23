package executor

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type cachedLogs struct {
	author uint64

	lastSeq uint64

	logs map[uint64][]*commonProto.OrderedLog

	logger external.Logger
}

func newCachedLogs(author uint64, logger external.Logger) *cachedLogs {
	return &cachedLogs{
		author: author,

		lastSeq: uint64(0),

		logs: make(map[uint64][]*commonProto.OrderedLog),

		logger: logger,
	}
}

func (cached *cachedLogs) write(exec *commonProto.ExecuteLogs) {
	if exec.Sequence <= cached.lastSeq {
		cached.logger.Warningf("Invalid sequence number, now last sequence %d", cached.lastSeq)
		return
	}

	cached.logger.Debugf("replica %d cached the logs of sequence %d", cached.author, exec.Sequence)
	cached.logs[exec.Sequence] = exec.OrderedLogs
}

func (cached *cachedLogs) read() []*commonProto.OrderedLog {
	sequence := cached.lastSeq+1
	cached.logger.Debugf("replica %d try to read the logs of sequence %d", cached.author, sequence)

	logs, ok := cached.logs[sequence]
	if !ok {
		return nil
	}

	delete(cached.logs, sequence)
	cached.lastSeq++
	return logs
}
