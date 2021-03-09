package api

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type Order interface {
	Basic

	Request(bid *commonProto.BatchId)

	Collect(msg *commonProto.OrderedMsg)
}
