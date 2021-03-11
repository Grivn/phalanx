package logmgr

import (
	"errors"
	"sync"

	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type logPool struct {
	id uint64

	sequence uint64
	recorder sync.Map

	logger external.Logger
}

func newLogPool(author uint64, logger external.Logger) *logPool {
	return &logPool{
		id: author,

		logger: logger,
	}
}

func (lp *logPool) save(msg *commonProto.OrderedMsg) {
	lp.logger.Infof("[LOG RECORD] sequence %d from replica %d", msg.Sequence, msg.Author)
	lp.recorder.Store(msg.BatchId.BatchHash, msg)
	lp.recorder.Store(msg.Sequence, msg)
}

func (lp *logPool) load(key interface{}) (*commonProto.OrderedMsg, error) {
	log, ok := lp.recorder.Load(key)
	if !ok {
		lp.logger.Errorf("cannot find current log")
		return nil, errors.New("cannot find log")
	}
	return log.(*commonProto.OrderedMsg), nil
}

func (lp *logPool) check(key interface{}) bool {
	_, ok := lp.recorder.Load(key)
	return ok
}

func (lp *logPool) remove(key interface{}) {
	lp.recorder.Delete(key)
}
