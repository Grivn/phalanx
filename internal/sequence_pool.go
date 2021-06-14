package internal

import "github.com/Grivn/phalanx/common/protos"

type SequencePool interface {
	InsertManager
	QCsManager
}

type InsertManager interface {
	// InsertQuorumCert could insert the quorum-cert into sync-tree for specific node.
	InsertQuorumCert(qc *protos.QuorumCert) error

	// InsertCommand could insert command into the sync-map.
	InsertCommand(command *protos.Command)
}

type QCsManager interface {
	// RestoreQCs is used to init the status of validator of QCs-manager.
	RestoreQCs()

	// PullQCs is used to pull the QCs from sync-tree to generate consensus proposal.
	PullQCs() ([]byte, error)

	// VerifyQCs is used to verify the QCs in qc-batch.
	// 1) we should have quorum QCs in such a batch.
	// 2) the qc should contain the specific command for it.
	// 3) the sequence number for qc should be matched with the local record for logs of replicas.
	// 4) the proof-certs should be valid.
	VerifyQCs(payload []byte) error

	StableQCs(payload []byte) error
}
