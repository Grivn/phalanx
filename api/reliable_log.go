package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type LogGenerator interface {
	Basic

	Generate(bid *commonProto.BatchId)

	Record(msg *commonProto.OrderedMsg)

	Ready(tag *commonProto.BinaryTag)
}
