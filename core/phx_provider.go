package phalanx

import "github.com/Grivn/phalanx/common/protos"

// SynchronousProvider is the phalanx service provider for (partial) synchronous consensus algorithm, such as PBFT or HS.
type SynchronousProvider interface {
	Communicator
	Generator
	Validator
	Executor
}

// Communicator is used to process messages from network.
type Communicator interface {
	// ProcessCommand is used to process the commands from clients.
	ProcessCommand(command *protos.Command)

	// ProcessConsensusMessage is used process the consensus messages from phalanx replica.
	ProcessConsensusMessage(message *protos.ConsensusMessage)
}

// Generator is used to generate essential messages.
type Generator interface {
	// MakePayload is used to generate payloads for bft consensus.
	MakePayload() ([]byte, error)
}

// Validator is used to verify the payload generated by phalanx module.
// there are 2 types of bft consensus algorithm for us:
// 1) chained-bft: before we turn into a new round, we need to generate a QC for the previous block,
//    it means that we could find a stable-height we generate the next proposal.
// 2) classic-bft: such as pbft, we cannot find a stable-height before we try to generate a checkpoint,
//    and we prefer to clean the persisted data after a checkpoint has been found.
// in order to use phalanx in these 2 types of bft consensus algorithm, we made the following interfaces.
type Validator interface {
	// Restore is used to restore data when we have found a timeout event in partial-synchronized bft consensus module.
	Restore()

	// Verify is used to verify the phalanx payload.
	// here, we would like to verify the validation of phalanx QCs, and record which seqNo has already been proposed.
	Verify(payload []byte) error

	// SetStable is used to set stable
	// here we would like to use it to control the order to process phalanx QCs.
	// when such a interface has returned error, a timeout event should be triggered.
	// 1) chained-bft: for each round we have generated a QC for chained-bft, we would like to use
	//    it to set phalanx stable status.
	// 2) classic-bft: for every time we are trying to execute a block, we would like to use it to
	//    set phalanx stable status.
	SetStable(payload []byte) error
}

// Executor is used to execute the phalanx objects.
type Executor interface {
	// Commit is used to commit the phalanx-QCBatch which has been verified by bft consensus.
	Commit(payload []byte) error
}
