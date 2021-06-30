package internal

import "github.com/Grivn/phalanx/common/protos"

type SequencePool interface {
	InsertManager
	PartialManager
}

type InsertManager interface {
	// InsertPartialOrder could insert the quorum-cert into sync-tree for specific node.
	InsertPartialOrder(pOrder *protos.PartialOrder) error

	// InsertCommand could insert command into the sync-map.
	InsertCommand(command *protos.Command)
}

type PartialManager interface {
	BecomeLeader()

	// RestorePartials is used to init the status of validator of POs-manager.
	RestorePartials()

	// PullPartials is used to pull the partial order from sync-tree to generate consensus proposal.
	PullPartials() (*protos.PartialOrderBatch, error)

	// VerifyPartials is used to verify the POs in qc-batch.
	// 1) we should have quorum POs in such a batch.
	// 2) the qc should contain the specific command for it.
	// 3) the sequence number for qc should be matched with the local record for logs of replicas.
	// 4) the proof-certs should be valid.
	VerifyPartials(pBatch *protos.PartialOrderBatch) error

	// SetStablePartials is used to process stable POs which have been verified by bft consensus.
	SetStablePartials(pBatch *protos.PartialOrderBatch) error
}
