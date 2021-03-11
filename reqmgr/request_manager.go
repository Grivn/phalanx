package reqmgr

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewRequestManager(n int, id uint64, replyC chan *commonProto.BatchId, logger external.Logger) api.RequestManager {
	return newRequestMgrImpl(n, id, replyC, logger)
}

func (rm *requestMgrImpl) Start() {
	rm.start()
}

func (rm *requestMgrImpl) Stop() {
	rm.stop()
}

func (rm *requestMgrImpl) Record(msg *commonProto.OrderedMsg) {
	rm.record(msg)
}
