package internal

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type MetaPool interface {
	api.Runner
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

//==================================== instance for meta pool =============================================

// ClientInstance is used to process commands info generated by specific client.
type ClientInstance interface {
	// Commit is used to notify the instance the committed command generated by current client.
	Commit(seqNo uint64) int

	// Append is used to notify the latest received command from current client.
	Append(command *protos.Command) int
}

// ReplicaInstance is used to process partial orders generated by each participant.
type ReplicaInstance interface {
	// GetHighOrder returns the highest partial order we have verified for current replica.
	GetHighOrder() *protos.PartialOrder

	// ReceivePreOrder is used to process the pre-order message from current replica.
	ReceivePreOrder(pre *protos.PreOrder) error

	// ReceivePartial is used to process the partial order message from current replica.
	ReceivePartial(pOrder *protos.PartialOrder) error
}

//================================== tracker for meta pool ========================================

// CommandTracker is used to record received commands.
type CommandTracker interface {
	RecordCommand(command *protos.Command)
	ReadCommand(digest string) *protos.Command
}

// PartialTracker is used to record received partial orders.
type PartialTracker interface {
	RecordPartial(pOrder *protos.PartialOrder)
	ReadPartial(idx types.QueryIndex) *protos.PartialOrder
	IsExist(idx types.QueryIndex) bool
}
