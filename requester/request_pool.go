package requester

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/requester/types"
)

type requestPool struct {
	id uint64

	sequence uint64
	recorder map[uint64]*commonProto.BatchId

	recvC  chan *commonProto.OrderedMsg
	replyC chan interface{}
	closeC chan bool

	logger external.Logger
}

func newRequestPool(id uint64, replyC chan interface{}, logger external.Logger) *requestPool {
	logger.Noticef("Init request pool for replica %d", id)
	return &requestPool{
		id: id,

		sequence: uint64(0),
		recorder: make(map[uint64]*commonProto.BatchId),

		recvC:  make(chan *commonProto.OrderedMsg),
		replyC: replyC,
		closeC: make(chan bool),

		logger: logger,
	}
}

func (rp *requestPool) start() {
	go rp.listener()
}

func (rp *requestPool) stop() {
	select {
	case <-rp.closeC:
	default:
		close(rp.closeC)
	}
}

func (rp *requestPool) record(msg *commonProto.OrderedMsg) {
	rp.recvC <- msg
}

func (rp *requestPool) listener() {
	for {
		select {
		case <-rp.closeC:
			rp.logger.Noticef("exist requestRecorderMgr listener for %d", rp.id)
			return
		case msg := <-rp.recvC:
			rp.processOrderedMsg(msg)
		}
	}
}

func (rp *requestPool) processOrderedMsg(msg *commonProto.OrderedMsg) {
	rp.logger.Infof("receive ordered request from replica %d, hash %s", msg.Author, msg.BatchId.BatchHash)

	if _, ok := rp.recorder[msg.Sequence]; ok {
		rp.logger.Warningf("already received batch for replica %d sequence %d", msg.Author, msg.Sequence)
		return
	}

	rp.recorder[msg.Sequence] = msg.BatchId
	for {
		bid, ok := rp.recorder[rp.sequence+1]
		if !ok {
			break
		}
		rp.sequence++
		rp.logger.Infof("propose batch id for replica %d sequence %d", rp.id, rp.sequence)

		event := types.ReplyEvent{
			EventType: types.ReqReplyBatchByOrder,
			Event:     bid,
		}

		rp.replyC <- event
		delete(rp.recorder, rp.sequence)
	}
}

