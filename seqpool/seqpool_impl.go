package seqpool

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/seqpool/types"
	"github.com/Grivn/phalanx/seqpool/utils"
	"sync"
)

type seqpoolImpl struct {
	author uint64

	requestMgr map[uint64]RequestManager

	logs sync.Map

	recvC chan *commonProto.OrderedMsg

	close chan bool

	logger external.Logger
}

func NewSequencePool(c types.Config) api.SequencePool {
	return newSeqpoolImpl(c)
}

func (sp *seqpoolImpl) Start() {
	sp.start()
}

func (sp *seqpoolImpl) Stop() {
	sp.stop()
}

func (sp *seqpoolImpl) Reset() {
	return
}

func (sp *seqpoolImpl) Record(msg *commonProto.OrderedMsg) {
	sp.record(msg)
}

func newSeqpoolImpl(c types.Config) *seqpoolImpl {
	reqMap := make(map[uint64]RequestManager)

	for index:= 0; index <c.N; index++ {
		id := uint64(index+1)
		reqMap[id] = utils.NewRequestMgr(id, c.ReplyC, c.Logger)
	}

	return &seqpoolImpl{
		author:     c.Author,
		requestMgr: reqMap,
		recvC:      make(chan *commonProto.OrderedMsg),
		close:      make(chan bool),
		logger:     c.Logger,
	}
}

func (sp *seqpoolImpl) start() {
	for _, re := range sp.requestMgr {
		re.Start()
	}

	go sp.listener()
}

func (sp *seqpoolImpl) stop() {
	close(sp.close)

	for _, re := range sp.requestMgr {
		re.Stop()
	}
}

func (sp *seqpoolImpl) record(msg *commonProto.OrderedMsg) {
	sp.recvC <- msg
}

func (sp *seqpoolImpl) listener() {
	for {
		select {
		case <-sp.close:
			sp.logger.Notice("exist sequence pool listener")
			return
		case msg := <-sp.recvC:
			sp.processMessage(msg)
		}
	}
}

func (sp *seqpoolImpl) processMessage(msg *commonProto.OrderedMsg) {
	switch msg.Type {
	case commonProto.OrderType_REQ:
		sp.processOrderedReq(msg)
	case commonProto.OrderType_LOG:
		sp.processOrderedLog(msg)
	default:
		sp.logger.Errorf("Invalid message type: code %d", msg.Type)
		return
	}
}

func (sp *seqpoolImpl) processOrderedReq(msg *commonProto.OrderedMsg) {
	sp.requestMgr[msg.Author].Update(msg)
}

func (sp *seqpoolImpl) processOrderedLog(msg *commonProto.OrderedMsg) {
	sp.logs.Store(msg.BatchId, msg)
}
