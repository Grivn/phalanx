package requester

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
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
	recvC commonTypes.RequesterRecvChan

	// sendC is the channel group which is used to send back information to other modules
	sendC commonTypes.RequesterSendChan

	// closeC is used to close the go-routine of request-pool
	closeC chan bool

	// logger is used to print logs
	logger external.Logger
}

func newRequestPool(author, id uint64, sendC commonTypes.RequesterSendChan, logger external.Logger) *requestPool {
	logger.Noticef("replica %d init request pool for replica %d", author, id)

	recvC := commonTypes.RequesterRecvChan{
		OrderedChan: make(chan *commonProto.OrderedReq),
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

func (rp *requestPool) record(req *commonProto.OrderedReq) {
	rp.recvC.OrderedChan <- req
}

func (rp *requestPool) listener() {
	for {
		select {
		case <-rp.closeC:
			rp.logger.Noticef("exist requestRecorderMgr listener for %d", rp.id)
			return
		case msg, ok := <-rp.recvC.OrderedChan:
			if !ok {
				continue
			}
			rp.processOrderedMsg(msg)
		}
	}
}

func (rp *requestPool) processOrderedMsg(req *commonProto.OrderedReq) {
	rp.logger.Infof("replica %d receive ordered request from replica %d, hash %s", rp.author, req.Author, req.BatchId.BatchHash)

	if _, ok := rp.recorder[req.Sequence]; ok {
		rp.logger.Warningf("already received batch for replica %d sequence %d", req.Author, req.Sequence)
		return
	}

	rp.recorder[req.Sequence] = req.BatchId
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
	rp.sendC.BatchIdChan <- bid
}
