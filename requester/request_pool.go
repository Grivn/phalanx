package requester

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type requestPool struct {
	// author is the local node's identifier
	author uint64

	// id indicates which node the current request-pool is maintained for
	id uint64

	// sequence is the preferred number for the next request
	sequence uint64

	// recorder is used to track the batch-info of node id
	recorder map[uint64]*commonProto.BatchId

	// recvC is the channel group which is used to receive events from other modules
	recvC recvChan

	// sendC is the channel group which is used to send back information to other modules
	sendC sendChan

	// closeC is used to close the go-routine of request-pool
	closeC chan bool

	// logger is used to print logs
	logger external.Logger
}

func newRequestPool(author, id uint64, bidC chan *commonProto.BatchId, logger external.Logger) *requestPool {
	logger.Noticef("replica %d init request pool for replica %d", author, id)

	recvC := recvChan{
		orderedChan: make(chan *commonProto.OrderedMsg),
	}

	sendC := sendChan{
		batchIdChan: bidC,
	}

	return &requestPool{
		author:   author,
		id:       id,
		sequence: uint64(0),
		recorder: make(map[uint64]*commonProto.BatchId),
		recvC:    recvC,
		sendC:    sendC,
		closeC:   make(chan bool),
		logger:   logger,
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
	rp.recvC.orderedChan <- msg
}

func (rp *requestPool) listener() {
	for {
		select {
		case <-rp.closeC:
			rp.logger.Noticef("exist requestRecorderMgr listener for %d", rp.id)
			return
		case msg, ok := <-rp.recvC.orderedChan:
			if !ok {
				continue
			}
			rp.processOrderedMsg(msg)
		}
	}
}

func (rp *requestPool) processOrderedMsg(msg *commonProto.OrderedMsg) {
	rp.logger.Infof("replica %d receive ordered request from replica %d, hash %s", rp.author, msg.Author, msg.BatchId.BatchHash)

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
		rp.logger.Infof("replica %d propose batch id for replica %d sequence %d", rp.author, rp.id, rp.sequence)

		go rp.sendSequentialBatch(bid)
		delete(rp.recorder, rp.sequence)
	}
}

func (rp *requestPool) sendSequentialBatch(bid *commonProto.BatchId) {
	rp.sendC.batchIdChan <- bid
}
