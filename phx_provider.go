package phalanx

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
)

// Provider is the phalanx service provider for all kinds of consensus algorithm, such as PBFT or HS.
type Provider interface {
	Runner
	Receiver
	Communicator
	Generator
	Executor

	// QueryMetrics returns the metrics info of phalanx.
	QueryMetrics() types.MetricsInfo
}

// Runner is used to control the modules which provide coroutine service.
type Runner interface {
	Run()
	Quit()
}

// Receiver is used to receive transactions and push them into phalanx memory pool.
type Receiver interface {
	// ReceiveTransaction is used to process transaction we have received.
	ReceiveTransaction(tx *protos.Transaction)
}

// Communicator is used to process messages from network.
type Communicator interface {
	// ReceiveCommand is used to process the commands from clients.
	ReceiveCommand(command *protos.Command)

	// ReceiveConsensusMessage is used process the consensus messages from phalanx replica.
	ReceiveConsensusMessage(message *protos.ConsensusMessage) error
}

// Generator is used to generate essential messages.
type Generator interface {
	// MakeProposal is used to generate phalanx proposal for consensus.
	MakeProposal() (*protos.PartialOrderBatch, error)
}

// Executor is used to execute the phalanx objects.
type Executor interface {
	// CommitProposal is used to commit the phalanx proposal which has been verified with consensus.
	CommitProposal(pBatch *protos.PartialOrderBatch) error
}
