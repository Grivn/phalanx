package logpool

import (
	"errors"
	"sync"

	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type logPoolImpl struct {
	id uint64

	sequence uint64
	recorder sync.Map

	logger external.Logger
}

func newLogPoolImpl(author uint64, logger external.Logger) *logPoolImpl {
	return &logPoolImpl{
		id: author,

		logger: logger,
	}
}

func (lp *logPoolImpl) save(msg *commonProto.OrderedMsg) {
	lp.logger.Infof("[LOG RECORD] sequence %d from replica %d", msg.Sequence, msg.Author)
	lp.recorder.Store(msg.BatchId.BatchHash, msg.Sequence)
	lp.recorder.Store(msg.Sequence, msg)
}

func (lp *logPoolImpl) load(key interface{}) (*commonProto.OrderedMsg, error) {
	log, ok := lp.recorder.Load(key)
	if !ok {
		return nil, errors.New("cannot find log")
	}
	return log.(*commonProto.OrderedMsg), nil
}

func (lp *logPoolImpl) check(key interface{}) bool {
	_, ok := lp.recorder.Load(key)
	return ok
}

func (lp *logPoolImpl) remove(key interface{}) {
	lp.recorder.Delete(key)
}
