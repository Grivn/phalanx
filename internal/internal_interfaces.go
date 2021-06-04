package internal

import "github.com/Grivn/phalanx/common/protos"

type LocalLog interface {
	ProcessCommand(command *protos.Command) (*protos.PreOrder, error)
	ProcessVote(vote *protos.Vote) (*protos.QuorumCert, error)
}

type RemoteLog interface {
	ProcessPreOrder(pre *protos.PreOrder)
	ProcessQC(qc *protos.QuorumCert)
}
