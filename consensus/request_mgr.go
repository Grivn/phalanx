package consensus

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type requestMgr struct {
	author uint64

	sequence uint64

	sequencePool api.LogPool

	auth *usigProxy

	sender *senderProxy

	recvC chan *commonProto.BatchId

	closeC chan bool

	logger external.Logger
}

func newRequestMgr(author uint64, sp api.LogPool, auth *usigProxy, sender *senderProxy, logger external.Logger) *requestMgr {
	return &requestMgr{
		author:       author,
		sequence:     uint64(0),
		sequencePool: sp,
		auth:         auth,
		sender:       sender,
		logger:       logger,
	}
}

func (rm *requestMgr) start() {
	go rm.listener()
}

func (rm *requestMgr) stop() {
	close(rm.closeC)
}

func (rm *requestMgr) propose(bid *commonProto.BatchId) {
	rm.recvC <- bid
}

func (rm *requestMgr) listener() {
	for {
		select {
		case <-rm.closeC:
			return
		case bid := <-rm.recvC:
			rm.generate(bid)
		}
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
	signed := rm.auth.generateSignedMsg(msg)
	rm.logger.Infof("[%d Generate] ordered req for seq %d batch %s", rm.author, rm.sequence, bid.BatchHash)

	rm.sequencePool.Record(msg)
	rm.sender.broadcast(signed)
}

func (rm *requestMgr) verify(signed *commonProto.SignedMsg) bool {
	if signed == nil {
		return false
	}

	return rm.auth.verifySignedMsg(signed)
}
