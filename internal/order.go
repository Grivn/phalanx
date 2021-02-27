package internal

import (
	commonBasic "github.com/Grivn/phalanx/common/basic"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type Order interface {
	commonBasic.Basic

	Request(bid *commonProto.BatchId)

	Collect(msg *commonProto.OrderedMsg)
}
