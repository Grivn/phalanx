package internal

import "github.com/Grivn/phalanx/common/protos"

// SyncTree is a b-tree which could deal with concurrency problems.
type SyncTree interface {
	// SubID returns which node current sync-tree is used for.
	SubID() uint64

	// Insert would insert the quorum cert into the b-tree of this node.
	Insert(qc *protos.QuorumCert)

	// Has is used to check if QC has been our in sync-tree.
	Has(qc *protos.QuorumCert) bool

	// Min returns the quorum-cert with the smallest sequence number for current node.
	Min() *protos.QuorumCert

	// PullMin returns the quorum-cert with the smallest sequence number for current node and remove it from b-tree.
	PullMin() *protos.QuorumCert

	// Delete is used to delete QC item.
	Delete(qc *protos.QuorumCert)
}
