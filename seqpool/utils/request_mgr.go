package utils

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/seqpool"
)

func NewRequestMgr(author uint64, replyC chan *commonProto.BatchId, logger external.Logger) seqpool.RequestManager {
	return newRequestMgr(author, replyC, logger)
}

func (rmi *requestMgrImpl) Start() {
	rmi.start()
}

func (rmi *requestMgrImpl) Stop() {
	rmi.stop()
}

func (rmi *requestMgrImpl) Reset() {
	return
}

func (rmi *requestMgrImpl) Save(msg *commonProto.OrderedMsg) {
	rmi.save(msg)
}

type requestMgrImpl struct {
	author   uint64
	sequence uint64
	recorder map[uint64]*commonProto.BatchId

	recvC  chan *commonProto.OrderedMsg
	replyC chan *commonProto.BatchId
	closeC chan bool

	logger external.Logger
}

func newRequestMgr(author uint64, replyC chan *commonProto.BatchId, logger external.Logger) *requestMgrImpl {
	return &requestMgrImpl{
		author:   author,
		sequence: uint64(0),
		recorder: make(map[uint64]*commonProto.BatchId),
		recvC:    make(chan *commonProto.OrderedMsg),
		replyC:   replyC,
		closeC:   make(chan bool),
		logger:   logger,
	}
}

func (rmi *requestMgrImpl) start() {
	go rmi.listener()
}

func (rmi *requestMgrImpl) stop() {
	close(rmi.closeC)
}

func (rmi *requestMgrImpl) save(msg *commonProto.OrderedMsg) {
	rmi.recvC <- msg
}

func (rmi *requestMgrImpl) listener() {
	for {
		select {
		case <-rmi.closeC:
			rmi.logger.Noticef("exist requestRecorderMgr listener for %d", rmi.author)
			return
		case msg := <-rmi.recvC:
			rmi.process(msg)
		}
	}
}

func (rmi *requestMgrImpl) process(msg *commonProto.OrderedMsg) {
	if _, ok := rmi.recorder[msg.Sequence]; ok {
		rmi.logger.Warningf("already received batch for replica %d sequence %d", msg.Author, msg.Sequence)
		return
	}

	rmi.recorder[msg.Sequence] = msg.BatchId
	for {
		bid, ok := rmi.recorder[rmi.sequence+1]
		if !ok {
			break
		}
		rmi.sequence++
		rmi.logger.Infof("propose batch id for replica %d sequence %d", rmi.author, rmi.sequence)
		rmi.replyC <- bid
		delete(rmi.recorder, rmi.sequence)
	}
}
