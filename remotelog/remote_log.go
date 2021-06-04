package reliablelog

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type remoteLog struct {
	// author is the current node's identifier
	author uint64

	subInstances map[uint64]SubInstance

	logger external.Logger
}

func NewReliableLog(n int, author uint64, logger external.Logger) *remoteLog {
	logger.Noticef("replica %d init log manager", author)

	return &remoteLog{
		author:    author,
		logger:    logger,
	}
}

func (rl *remoteLog) ProcessPreOrder(pre *commonProto.PreOrder) {
	err := rl.subInstances[pre.Author].ProcessPreOrder(pre)
	if err != nil {
		rl.logger.Errorf("%s", err)
	}
}

func (rl *remoteLog) ProcessQC(qc *commonProto.QuorumCert) {
	err := rl.subInstances[qc.Author()].ProcessQC(qc)
	if err != nil {
		rl.logger.Errorf("%s", err)
	}
}
