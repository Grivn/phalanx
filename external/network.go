package external

import commonProto "github.com/Grivn/phalanx/common/protos"

type Network interface {
	BroadcastPreOrder(pre *commonProto.PreOrder)

	SendVote(vote *commonProto.Vote, to uint64)

	BroadcastQC(qc *commonProto.QuorumCert)
}
