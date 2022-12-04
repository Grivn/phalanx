package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type MemoryPool interface {
	// ProcessLocalEvent is used to process the event received from local bus.
	ProcessLocalEvent(ev types.LocalEvent)

	// GenerateProposal generates proposal for consensus engine.
	// Here, we need to create phalanx proposal with each node highest order-attempt,
	// and make sure every order-attempt has a corresponding checkpoint.
	// If the order-attempt has not been verified, current node should send request
	// to generate checkpoint.
	GenerateProposal() (*protos.Proposal, error)

	// VerifyProposal verifies the proposal submitted from consensus engine,
	// and generates query stream if essential.
	VerifyProposal(proposal *protos.Proposal) (types.QueryStream, error)
}

type ConsensusEngine interface {
	Runner

	// ProcessConsensusMessage is used to process consensus messages.
	ProcessConsensusMessage(message *protos.ConsensusMessage)

	// ProcessLocalEvent is used to process local events.
	ProcessLocalEvent(event types.LocalEvent)
}

type SequencerInstance interface {
	// GetHighAttempt gets the latest legal order-attempt received from current sequencer.
	GetHighAttempt() *protos.OrderAttempt

	// Append is used to notify the latest received order-attempt from current sequencer and try to verify it.
	Append(attempt *protos.OrderAttempt)
}

//================================== tracker for meta pool ========================================

// CommandTracker is used to record received commands.
type CommandTracker interface {
	Record(command *protos.Command)
	Get(digest string) *protos.Command
}

// PartialTracker is used to record received partial orders.
type PartialTracker interface {
	Record(pOrder *protos.PartialOrder)
	Get(idx types.QueryIndex) *protos.PartialOrder
	IsExist(idx types.QueryIndex) bool
}

// AttemptTracker is used to record received order-attempts.
type AttemptTracker interface {
	Record(attempt *protos.OrderAttempt)
	Get(idx types.QueryIndex) *protos.OrderAttempt
}

// CheckpointTracker is used to record received checkpoint for order-attempts.
type CheckpointTracker interface {
	Record(checkpoint *protos.Checkpoint)
	Get(idx types.QueryIndex) *protos.Checkpoint
	IsExist(idx types.QueryIndex) bool
}
