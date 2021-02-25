package collector

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/order/collector/types"
	"github.com/Grivn/phalanx/order/external"

	logger "github.com/ultramesh/fancylogger"
)

type collectorImpl struct {
	author uint64

	validatorSet []uint64

	reqReceiver chan *commonProto.OrderedReq
	logReceiver chan *commonProto.OrderedLog

	reqCache map[msgID]*commonProto.OrderedReq
	logCache map[msgID]*commonProto.OrderedLog

	counterMap map[uint64]uint64

	close chan bool

	tee external.TEE

	sender external.Network

	sequencePool external.SequencePool

	logger *logger.Logger
}

func newCollectorImpl(c types.Config) *collectorImpl {
	return &collectorImpl{
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

func (ci *collectorImpl) start() {
	go ci.logListener()
	go ci.reqListener()
}

func (ci *collectorImpl) stop() {
	close(ci.close)
}

func (ci *collectorImpl) reqListener() {
	for {
		select {
		case <-ci.close:
			ci.logger.Notice("exist request listener")
			return
		case req := <-ci.reqReceiver:
			ci.processOrderedReq(req)
		}
	}
}

func (ci *collectorImpl) logListener() {
	for {
		select {
		case <-ci.close:
			ci.logger.Notice("exist log listener")
			return
		case log := <-ci.logReceiver:
			ci.processOrderedLog(log)
		}
	}
}

func (ci *collectorImpl) processOrderedReq(req *commonProto.OrderedReq) {
	if req.Author == ci.author {
		ci.logger.Infof("Replica %d received ordered request from self", ci.author)
		return
	}
}

func (ci *collectorImpl) processOrderedLog(log *commonProto.OrderedLog) {

}
