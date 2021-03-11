package phalanx

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	requestMgr2 "github.com/Grivn/phalanx/reqmgr"
)

type requestMgr struct {
	N int

	author uint64

	sequence uint64

	pools map[uint64]api.RequestPool

	sender *senderProxy

	recvC chan interface{}

	closeC chan bool

	logger external.Logger
}

func newRequestMgr(n int, author uint64, replyC chan *commonProto.BatchId, sender *senderProxy, logger external.Logger) *requestMgr {
	logger.Noticef("Init the request manager of replica %d", author)

	initSeq := uint64(0)

	rps := make(map[uint64]api.RequestPool)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		rp := requestMgr2.NewRequestPool(id, replyC, logger)
		rps[id] = rp
	}

	return &requestMgr{
		N:        n,
		author:   author,
		sequence: initSeq,
		pools:    rps,
		sender:   sender,
		logger:   logger,
	}
}

func (rm *requestMgr) start() {
	go rm.listener()
}

func (rm *requestMgr) stop() {
	close(rm.closeC)
}

func (rm *requestMgr) propose(event interface{}) {
	rm.recvC <- event
}

func (rm *requestMgr) listener() {
	for {
		select {
		case <-rm.closeC:
			rm.logger.Notice("exist request manager listener")
			return
		case ev := <-rm.recvC:
			rm.processRequestEvents(ev)
		}
	}
}

func (rm *requestMgr) processRequestEvents(event interface{}) {
	switch e := event.(type) {
	case *commonProto.BatchId:
		rm.generate(e)
	case *commonProto.OrderedMsg:
		rm.pools[e.Author].Record(e)
	}
}

func (rm *requestMgr) generate(bid *commonProto.BatchId) {
	if bid == nil {
		rm.logger.Warningf("[%d Warning] received a nil batch id", rm.author)
		return
	}

	rm.sequence++
	msg := &commonProto.OrderedMsg{
		Type:     commonProto.OrderType_REQ,
		Author:   rm.author,
		Sequence: rm.sequence,
		BatchId:  bid,
	}

	rm.logger.Infof("[%d Generate] ordered req for seq %d batch %s", rm.author, rm.sequence, bid.BatchHash)

	rm.pools[rm.author].Record(msg)
	rm.sender.broadcast(msg)
}
