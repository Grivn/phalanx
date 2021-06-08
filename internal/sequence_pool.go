package internal

import "github.com/Grivn/phalanx/common/protos"

type SequencePool interface {
	// InsertQuorumCert could insert the quorum-cert into sync-tree for specific node.
	InsertQuorumCert(qc *protos.QuorumCert)

	// InsertCommand could insert command into the sync-map.
	InsertCommand(command *protos.Command)

	// VerifyQCs is used to verify the QCs in qc-batch.
	// 1) we should have quorum QCs in such a batch.
	// 2) the qc should contain the specific command for it.
	// 3) the sequence number for qc should be matched with the local record for logs of replicas.
	// 4) the proof-certs should be valid.
	VerifyQCs(payload []byte) error
	// LockQCs is used to delete the stable-QCs which have been proposed by other nodes.
	LockQCs(payload []byte) error
	// PullQCs is used to pull the QCs from sync-tree to generate consensus proposal.
	PullQCs() ([]byte, error)
}
