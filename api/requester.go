package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type Requester interface {
	Basic

	Generate(bid *commonProto.BatchId)

	Record(msg *commonProto.OrderedMsg)
}
