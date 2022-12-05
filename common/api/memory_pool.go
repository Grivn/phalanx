package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type MemoryPool interface {
	MemoryRelay
	MemoryProposer
	MemoryReader
}

type MemoryRelay interface {
	ProcessCommand(command *protos.Command)
	ProcessOrderAttempt(attempt *protos.OrderAttempt)
	ProcessConsensusMessage(consensusMessage *protos.ConsensusMessage)
}

type MemoryReader interface {
	// ReadCommand reads raw command from meta pool.
	ReadCommand(commandD string) *protos.Command

	// ReadOrderAttempts reads order attempts according to query stream.
	ReadOrderAttempts(qStream types.QueryStream) []*protos.OrderAttempt
}

type MemoryProposer interface {
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
