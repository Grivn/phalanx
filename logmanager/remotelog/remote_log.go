package reliablelog

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type remoteLog struct {
	// author is the current node's identifier.
	author uint64

	// subInstances is the module for us to process consensus messages for participates.
	subInstances map[uint64]SubInstance

	// logger is used to print logs
	logger external.Logger
}

func NewRemoteLog(n int, author uint64, logger external.Logger) *remoteLog {
	logger.Noticef("replica %d init log manager", author)

	return &remoteLog{
		author:    author,
		logger:    logger,
	}
}

// ProcessPreOrder is used as a proxy for remote-log module to process pre-order messages.
func (remote *remoteLog) ProcessPreOrder(pre *protos.PreOrder) {
	err := remote.subInstances[pre.Author].ProcessPreOrder(pre)
	if err != nil {
		remote.logger.Errorf("%s", err)
	}
}

// ProcessQC is used as a proxy for remote-log module to process QC messages.
func (remote *remoteLog) ProcessQC(qc *protos.QuorumCert) {
	err := remote.subInstances[qc.Author()].ProcessQC(qc)
	if err != nil {
		remote.logger.Errorf("%s", err)
	}
}
