package internal

import "github.com/Grivn/phalanx/common/protos"

type LocalLog interface {
	// ProcessCommand is used to process command received from clients.
	// We would like to assign a sequence number for such a command and generate a pre-order message.
	ProcessCommand(command *protos.Command) (*protos.PreOrder, error)

	// ProcessVote is used to process the vote message from others.
	// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
	ProcessVote(vote *protos.Vote) (*protos.QuorumCert, error)
}

type RemoteLog interface {
	// ProcessPreOrder is used as a proxy for remote-log module to process pre-order messages.
	ProcessPreOrder(pre *protos.PreOrder)

	// ProcessQC is used as a proxy for remote-log module to process QC messages.
	ProcessQC(qc *protos.QuorumCert)
}
