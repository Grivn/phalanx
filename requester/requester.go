package requester

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

func NewRequester(n int, id uint64, sendC commonTypes.RequesterSendChan, network external.Network, logger external.Logger) internal.Requester {
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
