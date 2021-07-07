package phalanx

import "github.com/Grivn/phalanx/common/protos"

// Provider is the phalanx service provider for all kinds of consensus algorithm, such as PBFT or HS.
type Provider interface {
	Communicator
	Generator
	Validator
	Executor
	TestReceiver
}

type TestReceiver interface {
	ProcessTransaction(tx *protos.Transaction)
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
	// MakeProposal is used to generate proposals for bft consensus.
	MakeProposal(priori *protos.PartialOrderBatch) (*protos.PartialOrderBatch, error)
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

	// Verify is used to verify the phalanx proposal.
	// here, we would like to verify the validation of partial orders, and make sure the sequence number hasn't proposed before
	// after that we should record the sequence number which has already been proposed and update specific status.
	Verify(pBatch *protos.PartialOrderBatch) error

	// SetStable is used to set stable
	// call set stable interface when bft consensus reach a stable status, which means such a block won't be withdrawn,
	// in order to make sure that current block or quorum cert could be proposed by order.
	//
	// if set stable failed, it means the stable partial orders are not proposed by order, just call restore interface for
	// recovery and trigger leader election if the bft protocol we use is not a rotating leader protocol:
	//
	// 1) chained bft: such as HotStuff, at 2-chained we could trigger set stable.
	// 2) classic bft: such as PBFT, set stable when before we try to commit bft block, we could generate a nil phalanx
	//                 proposal to replace the invalid one and trigger view-change for fixed leader protocol.
	SetStable(pBatch *protos.PartialOrderBatch) error
}

// Executor is used to execute the phalanx objects.
type Executor interface {
	// Commit is used to commit the phalanx-QCBatch which has been verified by bft consensus.
	Commit(pBatch *protos.PartialOrderBatch) error
}
