package requester

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/requester/types"
)

func NewRequester(n int, id uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) api.Requester {
	return newRequesterImpl(n, id, replyC, network, logger)
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

func (r *requesterImpl) Record(msg *commonProto.OrderedMsg) {
	r.record(msg)
}
