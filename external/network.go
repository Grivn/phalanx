package external

import commonProto "github.com/Grivn/phalanx/common/protos"

type Network interface {
	BroadcastBatch(batch *commonProto.Batch)

	BroadcastProposal(proposal *commonProto.Proposal)

	BroadcastPreOrder(pre *commonProto.PreOrder)

	SendVote(vote *commonProto.Vote, to uint64)

	BroadcastOrder(order *commonProto.Order)

	BroadcastReq(req *commonProto.OrderedReq)

	BroadcastLog(log *commonProto.OrderedLog)

	BroadcastAck(ack *commonProto.OrderedAck)
}
