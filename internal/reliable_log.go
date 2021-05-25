package internal

import commonProto "github.com/Grivn/phalanx/common/protos"

type ReliableLog interface {
	Basic

	Generate(bid *commonProto.BatchId)

	RecordLog(log *commonProto.OrderedLog)

	RecordAck(ack *commonProto.OrderedAck)
}
