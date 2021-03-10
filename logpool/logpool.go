package logpool

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewLogPool(id uint64, logger external.Logger) api.LogPool {
	return newLogPoolImpl(id, logger)
}

func (lp *logPoolImpl) ID() uint64 {
	return lp.id
}

func (lp *logPoolImpl) Save(msg *commonProto.OrderedMsg) {
	lp.save(msg)
}

func (lp *logPoolImpl) Load(key interface{}) (*commonProto.OrderedMsg, error) {
	return lp.load(key)
}

func (lp *logPoolImpl) Check(key interface{}) bool {
	return lp.check(key)
}

func (lp *logPoolImpl) Remove(key interface{}) {
	lp.remove(key)
}
