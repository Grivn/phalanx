package internal

import "github.com/Grivn/phalanx/common/protos"

type LogManager interface {
	Run()
	Committed(author uint64, seqNo uint64)
	LocalLog
	RemoteLog
}

type LocalLog interface {
	// ProcessCommand is used to process command received from clients.
	// We would like to assign a sequence number for such a command and generate a pre-order message.
	ProcessCommand(command *protos.PCommand) error

	// ProcessVote is used to process the vote message from others.
	// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
	ProcessVote(vote *protos.PVote) error
}

type RemoteLog interface {
	// ProcessPreOrder is used to process pre-order messages.
	// We should make sure that we have never received a pre-order/order message
	// whose sequence number is the same as it yet, and we would like to generate a
	// vote message for it if it's legal for us.
	ProcessPreOrder(pre *protos.PreOrder) error

	// ProcessPartial is used to process quorum-cert messages.
	// A valid quorum-cert message, which has a series of valid signature which has reached quorum size,
	// could advance the sequence counter. We should record the advanced counter and put the info of
	// order message into the sequential-pool.
	ProcessPartial(pOrder *protos.PartialOrder) error
}
