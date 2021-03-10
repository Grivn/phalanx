package reqpool

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type requestPoolImpl struct {
	id uint64

	sequence uint64
	recorder map[uint64]*commonProto.BatchId

	recvC  chan *commonProto.OrderedMsg
	replyC chan *commonProto.BatchId
	closeC chan bool

	logger external.Logger
}

func newRequestPoolImpl(id uint64, replyC chan *commonProto.BatchId, logger external.Logger) *requestPoolImpl {
	logger.Noticef("Init request pool for replica %d", id)
	return &requestPoolImpl{
		id: id,

		sequence: uint64(0),
		recorder: make(map[uint64]*commonProto.BatchId),

		recvC:  make(chan *commonProto.OrderedMsg),
		replyC: replyC,
		closeC: make(chan bool),

		logger: logger,
	}
}

func (rp *requestPoolImpl) start() {
	go rp.listener()
}

func (rp *requestPoolImpl) stop() {
	select {
	case <-rp.closeC:
	default:
		close(rp.closeC)
	}
}

func (rp *requestPoolImpl) record(msg *commonProto.OrderedMsg) {
	rp.recvC <- msg
}

func (rp *requestPoolImpl) listener() {
	for {
		select {
		case <-rp.closeC:
			rp.logger.Noticef("exist requestRecorderMgr listener for %d", rp.id)
			return
		case msg := <-rp.recvC:
			rp.process(msg)
		}
	}
}

func (rp *requestPoolImpl) process(msg *commonProto.OrderedMsg) {
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
		rp.replyC <- bid
		delete(rp.recorder, rp.sequence)
	}
}

