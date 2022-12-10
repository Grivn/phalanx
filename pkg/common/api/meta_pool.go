package api

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
)

type MetaPool interface {
	Runner
	LogManager
	MetaReader
	MetaCommitter
	MetaConsensus
}

type LogManager interface {
	LocalLog
	RemoteLog
}

type LocalLog interface {
	// ProcessCommand is used to process command received from clients.
	// We would like to assign a sequence number for such a command and generate a pre-order message.
	ProcessCommand(command *protos.Command)

	// ProcessVote is used to process the vote message from others.
	// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
	ProcessVote(vote *protos.Vote) error
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

type MetaReader interface {
	// ReadCommand reads raw command from meta pool.
	ReadCommand(commandD string) *protos.Command

	// ReadPartials reads partial orders according to query stream.
	ReadPartials(qStream types.QueryStream) []*protos.PartialOrder
}

type MetaCommitter interface {
	// Committed notifies meta pool the committed command info.
	Committed(author uint64, seqNo uint64)
}

type MetaConsensus interface {
	// GenerateProposal generates proposal for total consensus processor.
	GenerateProposal() (*protos.PartialOrderBatch, error)

	// VerifyProposal verifies the proposal consented by consensus processor
	// and generate query stream for partial orders if essential.
	VerifyProposal(batch *protos.PartialOrderBatch) (types.QueryStream, error)
}
