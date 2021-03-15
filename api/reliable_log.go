package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type ReliableLog interface {
	Basic

	Generate(bid *commonProto.BatchId)

	Record(msg *commonProto.SignedMsg)

	Ready(tag *commonProto.BinaryTag)
}
