package utils

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"sync"
)

type logMgrImpl struct {
	// recorder for author
	author uint64

	sequence uint64

	seqRecorder sync.Map

	logRecorder sync.Map

	logger external.Logger
}

type msgID struct {
	author uint64
	hash   string
}

func newLogMgrImpl(author uint64, logger external.Logger) *logMgrImpl {
	return &logMgrImpl{
		author: author,
		logger: logger,
	}
}

func (lmi *logMgrImpl) save(msg *commonProto.OrderedMsg) {
	lmi.logger.Infof("[LOG RECORD] sequence %d from replica %d", msg.Sequence, msg.Author)
	lmi.seqRecorder.Store(msg.BatchId.BatchHash, msg.Sequence)
	lmi.logRecorder.Store(msg.Sequence, msg)
}
