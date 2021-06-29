package executor

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type CollectionRule interface {
	TryCollection(qc *protos.QuorumCert) bool
}

type collectionRule struct {
	// quorum indicates the legal size for bft.
	quorum int

	// cmdCounter is used to calculate the 
	cmdCounter map[string]int
}

func newCollectRule(n int) *collectionRule {
	return &collectionRule{
		quorum:     types.CalculateQuorum(n),
		cmdCounter: make(map[string]int),
	}
}

func (cr *collectionRule) TryCollection(qc *protos.QuorumCert) bool {
	if cr.cmdCounter[qc.CommandDigest()] == cr.quorum {
		return false
	}

	cr.cmdCounter[qc.CommandDigest()]++
	return true
}
