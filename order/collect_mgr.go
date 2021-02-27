package order

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	phalanxInternal "github.com/Grivn/phalanx/internal"

	"github.com/Grivn/phalanx/order/types"
)

type collectorMgr struct {
	author uint64

	validatorSet []uint64

	recvC chan *commonProto.OrderedMsg

	reqCache map[msgID]*commonProto.OrderedMsg
	logCache map[msgID]*commonProto.OrderedLog

	counterMap map[uint64]uint64

	close chan bool

	tee phalanxInternal.USIG

	sender external.Network

	sequencePool phalanxInternal.SequencePool

	logger external.Logger
}

func newCollectMgr(c types.ColConfig) *collectorMgr {
	return &collectorMgr{
		author: c.Author,

		reqReceiver: c.ReqC,
		logReceiver: c.LogC,
		close:       make(chan bool),

		tee:          c.TEE,
		sender:       c.Sender,
		sequencePool: c.SequencePool,

		logger: c.Logger,
	}
}

func (cm *collectorMgr) start() {
	go cm.logListener()
	go cm.reqListener()
}

func (cm *collectorMgr) stop() {
	close(cm.close)
}

func (cm *collectorMgr) reqListener() {
	for {
		select {
		case <-cm.close:
			cm.logger.Notice("exist request listener")
			return
		case req := <-cm.reqReceiver:
			cm.processOrderedReq(req)
		}
	}
}

func (cm *collectorMgr) logListener() {
	for {
		select {
		case <-cm.close:
			cm.logger.Notice("exist log listener")
			return
		case log := <-cm.logReceiver:
			cm.processOrderedLog(log)
		}
	}
}

func (cm *collectorMgr) processOrderedReq(req *commonProto.OrderedReq) {
	if req.Author == cm.author {
		cm.logger.Infof("Replica %d received ordered request from self", cm.author)
		return
	}
}

func (cm *collectorMgr) processOrderedLog(log *commonProto.OrderedLog) {

}
