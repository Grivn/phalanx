package requester

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/requester/types"
)

type requesterImpl struct {
	n int

	author uint64

	sequence uint64

	pool map[uint64]*requestPool

	replyC chan types.ReplyEvent

	sender *senderProxy

	logger external.Logger
}

func newRequesterImpl(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) *requesterImpl {
	logger.Noticef("[INIT] replica %d init request manager, cluster amount %d", author, n)
	rps := make(map[uint64]*requestPool)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		rp := newRequestPool(id, replyC, logger)
		rps[id] = rp
	}

	return &requesterImpl{
		n:      n,
		author: author,
		pool:   rps,
		replyC: replyC,
		sender: newSenderProxy(author, network),
		logger: logger,
	}
}

func (r *requesterImpl) start() {
	for _, pool := range r.pool {
		pool.start()
	}
}

func (r *requesterImpl) stop() {
	for _, pool := range r.pool {
		pool.stop()
	}
}

func (r *requesterImpl) generate(bid *commonProto.BatchId) {
	if bid == nil {
		r.logger.Warningf("[%d Warning] received a nil batch id", r.author)
		return
	}

	r.sequence++
	msg := &commonProto.OrderedMsg{
		Type:     commonProto.OrderType_REQ,
		Author:   r.author,
		Sequence: r.sequence,
		BatchId:  bid,
	}

	r.logger.Infof("[%d Generate] ordered req for seq %d batch %s", r.author, r.sequence, bid.BatchHash)
	r.sender.broadcast(msg)
	r.record(msg)
}

func (r *requesterImpl) record(msg *commonProto.OrderedMsg) {
	r.pool[msg.Author].record(msg)
}
