package reliablelog

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type reliableLogImpl struct {
	// author is the current node's identifier
	author uint64

	subInstances map[uint64]SubInstance

	generator *orderMgr

	logger external.Logger
}

func newReliableLogImpl(n int, author uint64, sendC commonTypes.ReliableSendChan, network external.Network, logger external.Logger) *reliableLogImpl {
	logger.Noticef("replica %d init log manager", author)

	return &reliableLogImpl{
		author:    author,
		generator: newOrderMgr(author, logger),
		logger:    logger,
	}
}
