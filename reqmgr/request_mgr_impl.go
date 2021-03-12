package reqmgr

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type requestMgrImpl struct {
	n int

	author uint64

	sequence uint64

	pool map[uint64]*requestPool

	replyC chan interface{}

	sender *senderProxy

	logger external.Logger
}

func newRequestMgrImpl(n int, author uint64, replyC chan interface{}, network external.Network, logger external.Logger) *requestMgrImpl {
	logger.Noticef("[INIT] replica %d init request manager, cluster amount %d", author, n)
	rps := make(map[uint64]*requestPool)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		rp := newRequestPool(id, replyC, logger)
		rps[id] = rp
	}

	return &requestMgrImpl{
		n:      n,
		author: author,
		pool:   rps,
		replyC: replyC,
		sender: newSenderProxy(author, network),
		logger: logger,
	}
}

func (rm *requestMgrImpl) start() {
	for _, pool := range rm.pool {
		pool.start()
	}
}

func (rm *requestMgrImpl) stop() {
	for _, pool := range rm.pool {
		pool.stop()
	}
}

func (rm *requestMgrImpl) generate(bid *commonProto.BatchId) {
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
	rm.sender.broadcast(msg)
	rm.record(msg)
}

func (rm *requestMgrImpl) record(msg *commonProto.OrderedMsg) {
	rm.pool[msg.Author].record(msg)
}
