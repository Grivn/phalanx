package requester

import (
	"github.com/Grivn/phalanx/api"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewRequester(n int, id uint64, sendC commonTypes.RequesterSendChan, network external.Network, logger external.Logger) api.Requester {
	return newRequesterImpl(n, id, sendC, network, logger)
}

func (r *requesterImpl) Start() {
	r.start()
}

func (r *requesterImpl) Stop() {
	r.stop()
}

func (r *requesterImpl) Generate(bid *commonProto.BatchId) {
	r.generate(bid)
}

func (r *requesterImpl) Record(msg *commonProto.OrderedReq) {
	r.record(msg)
}
