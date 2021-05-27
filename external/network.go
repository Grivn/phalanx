package external

import commonProto "github.com/Grivn/phalanx/common/protos"

type Network interface {
	BroadcastBatch(batch *commonProto.Batch)

	BroadcastProposal(proposal *commonProto.Proposal)

	BroadcastReq(req *commonProto.OrderedReq)

	BroadcastLog(log *commonProto.OrderedLog)

	BroadcastAck(ack *commonProto.OrderedAck)
}
