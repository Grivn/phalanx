package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type MetaPool interface {
	// ProcessCommand is used to process command it has received.
	ProcessCommand(command *protos.Command)

	// GenerateProposal generates proposal for consensus engine.
	// Here, we need to create phalanx proposal with each node highest order-attempt,
	// and make sure every order-attempt has a corresponding checkpoint.
	// If the order-attempt has not been verified, current node should send request
	// to generate checkpoint.
	GenerateProposal() (*protos.Proposal, error)

	// VerifyProposal verifies the proposal submitted from consensus engine,
	// and generates query stream if essential.
	VerifyProposal(proposal *protos.Proposal) (types.QueryStream, error)

	CheckpointProcessor
}

type CheckpointProcessor interface {
	// GetCheckpoint is used to get the target checkpoint.
	GetCheckpoint(idx types.QueryIndex)

	// ProcessCheckpointRequest is used to process checkpoint request.
	ProcessCheckpointRequest(request *protos.CheckpointRequest)

	// ProcessCheckpointVote is used to process checkpoint vote.
	ProcessCheckpointVote(vote *protos.CheckpointVote)

	// ProcessCheckpoint is used to process checkpoint.
	ProcessCheckpoint(checkpoint *protos.Checkpoint)
}

type SequencerInstance interface {
	// GetHighAttempt gets the latest legal order-attempt received from current sequencer.
	GetHighAttempt() *protos.OrderAttempt

	// Append is used to notify the latest received order-attempt from current sequencer and try to verify it.
	Append(attempt *protos.OrderAttempt)
}

// Relay (module for experiment) is used to relay the commands from specific client with pre-defined ordering strategy.
type Relay interface {
	// Append is used to notify the latest received command from current client.
	Append(command *protos.Command) int

	// Commit is used to notify the instance the committed command generated by current client.
	Commit(seqNo uint64) int
}
