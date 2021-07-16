package internal

import "github.com/Grivn/phalanx/common/protos"

type SequencePool interface {
	InsertManager
	PartialManager
}

type InsertManager interface {
	// InsertPartialOrder could insert the partial order into b-tree for specific node.
	InsertPartialOrder(pOrder *protos.PartialOrder) error

	// InsertCommand could insert command into the command recorder.
	InsertCommand(command *protos.PCommand)
}

type PartialManager interface {
	// RestorePartials is used to init the status of validator of partial order manager.
	RestorePartials()

	// PullPartials is used to pull the partial order from sync-tree to generate consensus proposal.
	PullPartials(priori *protos.PartialOrderBatch) (*protos.PartialOrderBatch, error)

	// VerifyPartials is used to verify the partial order batch.
	// 1) the command a partial order refer to should be combined in batch.
	// 2) the partial order should not be proposed before.
	// 3) the proof-certs should be valid.
	VerifyPartials(pBatch *protos.PartialOrderBatch) error

	// SetStablePartials is used to process stable partial order which have been verified by bft consensus.
	SetStablePartials(pBatch *protos.PartialOrderBatch) error
}
