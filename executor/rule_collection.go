package executor

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type collectionRule struct {
	// quorum indicates the legal size for bft.
	quorum int

	//
	recorder *commandRecorder
}

func newCollectRule(n int) *collectionRule {
	return &collectionRule{
		quorum:     types.CalculateQuorum(n),
	}
}

func (cr *collectionRule) collect(qc *protos.QuorumCert) bool {

	return true
}
