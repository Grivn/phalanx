package order

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/order/types"
)

type requestMgr struct {
	author uint64

	sequence uint64

	recvC chan *commonProto.BatchId

	close chan bool

	sender external.Network

	usig internal.USIG

	sequencePool internal.SequencePool

	logger external.Logger
}

func newRequestMgr(config types.ReqConfig) *requestMgr {
	return &requestMgr{
		author:       config.Author,
		sequence:     uint64(0),
		recvC:        make(chan *commonProto.BatchId, 1000),
		close:        make(chan bool),
		sender:       config.Network,
		usig:         config.USIG,
		sequencePool: config.SeqPool,
		logger:       config.Logger,
	}
}

func (rm *requestMgr) start() {
	go rm.listener()
}

func (rm *requestMgr) stop() {
	close(rm.close)
}

func (rm *requestMgr) request(bid *commonProto.BatchId) {
	rm.recvC <- bid
}

func (rm *requestMgr) listener() {
	for {
		select {
		case bid := <-rm.recvC:
			rm.generateOrderedRequest(bid)
		case <-rm.close:
			rm.logger.Notice("exist requester listener")
			return
		}
	}
}

func (rm *requestMgr) generateOrderedRequest(bid *commonProto.BatchId) {
	if bid == nil {
		return
	}


}
