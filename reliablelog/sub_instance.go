package reliablelog

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type subInstance struct {
	// author is the local node's identifier
	author uint64

	// id indicates which node the current request-pool is maintained for
	id uint64

	// sequence is the preferred number for the next request
	sequence uint64

	// recorder is used to track the pre-order
	recorder map[uint64]*commonProto.PreOrder

	// preC is used to receive the pre-order message from others
	preC chan *commonProto.PreOrder

	// sender is used to send votes to others
	sender external.Network

	// closeC is used to close the go-routine of request-pool
	closeC chan bool

	// logger is used to print logs
	logger external.Logger
}

func newSubInstance(author, id uint64, preC chan *commonProto.PreOrder, sender external.Network, logger external.Logger) *subInstance {
	logger.Noticef("replica %d init the sub instance of order for replica %d", author, id)
	return &subInstance{
		author:   author,
		id:       id,
		sequence: uint64(0),
		recorder: make(map[uint64]*commonProto.PreOrder),
		preC:     preC,
		closeC:   make(chan bool),
		sender:   sender,
		logger:   logger,
	}
}

func (si *subInstance) processPreOrder(pre *commonProto.PreOrder) {
	si.logger.Infof("replica %d received a pre-order message from replica %d, hash %s", si.author, pre.Author, pre.Digest)

	if _, ok := si.recorder[pre.Sequence]; !ok {
		si.logger.Warningf("replica %d has already received pre-order from replica %d on sequence %d", si.author, pre.Author, pre.Sequence)
		return
	}

	si.recorder[pre.Sequence] = pre
	for {
		_, ok := si.recorder[si.sequence+1]
		if !ok {
			break
		}
		si.sequence++

		// todo process pre-order message
		delete(si.recorder, si.sequence)
	}
}
