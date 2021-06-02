package internal

import "github.com/Grivn/phalanx/common/protos"

type SequencePool interface {
	InsertQuorumCert(cert *protos.QuorumCert)
}
