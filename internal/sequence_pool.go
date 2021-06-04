package internal

import "github.com/Grivn/phalanx/common/protos"

type SequencePool interface {
	InsertManager
}

type InsertManager interface {
	InsertQuorumCert(cert *protos.QuorumCert)
}
