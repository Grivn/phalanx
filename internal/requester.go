package internal

import commonProto "github.com/Grivn/phalanx/common/protos"

type Requester interface {
	Basic

	Generate(bid *commonProto.BatchId)

	Record(msg *commonProto.OrderedReq)
}
