package reqpool

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewRequestPool(id uint64, replyC chan *commonProto.BatchId, logger external.Logger) api.RequestPool {
	return newRequestPoolImpl(id, replyC, logger)
}

func (rp *requestPoolImpl) Start() {
	rp.start()
}

func (rp *requestPoolImpl) Stop() {
	rp.stop()
}

// ID returns which replica the requests in current pool belong to
func (rp *requestPoolImpl) ID() uint64 {
	return rp.id
}

func (rp *requestPoolImpl) Record(msg *commonProto.OrderedMsg) {
	rp.record(msg)
}
