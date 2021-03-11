package phalanx

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	txpoolTypes "github.com/Grivn/phalanx/txpool/types"
)

type phalanxImpl struct {
	author uint64

	txpool  api.TxPool
	requestMgr *requestMgr
	logMgr *logMgrImpl
	logpool api.LogPool

	recvC chan interface{}
	txpoolC  chan txpoolTypes.ReplyEvent
	reqpoolC chan *commonProto.BatchId
	closeC   chan bool

	logger external.Logger
}

func newPhalanxImpl() *phalanxImpl {
	return &phalanxImpl{}
}

func (phi *phalanxImpl) postTxs(txs []*commonProto.Transaction) {
	phi.logger.Infof("Replica %d received transferred txs from api", phi.author)
	for _, tx := range txs {
		phi.txpool.PostTx(tx)
	}
}

func (phi *phalanxImpl) propose(event interface{}) {
	phi.recvC <- event
}

func (phi *phalanxImpl) listenTxPool() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Noticef("exist main listener for phalanx")
			return
		case event := <-phi.recvC:
			phi.dispatchEvents(event)
		}
	}
}

func (phi *phalanxImpl) dispatchEvents(event interface{}) {
	switch e := event.(type) {
	case txpoolTypes.ReplyEvent:
		phi.processTxPoolEvents(e)
	case *commonProto.OrderedMsg:
		phi.processOrderedMsg(e)
	case *commonProto.BatchId:
		phi.logMgr.propose(e)
	}
}

func (phi *phalanxImpl) processTxPoolEvents(event txpoolTypes.ReplyEvent) {
	switch event.EventType {
	case txpoolTypes.ReplyGenerateBatchEvent:
		batch := event.Event.(*commonProto.Batch)
		phi.requestMgr.generate(batch.BatchId)
	case txpoolTypes.ReplyLoadBatchEvent:
	case txpoolTypes.ReplyMissingBatchEvent:
	default:
		phi.logger.Warningf("Invalid event type: code %d", event.EventType)
	}
}

func (phi *phalanxImpl) processOrderedMsg(msg *commonProto.OrderedMsg) {
	switch msg.Type {
	case commonProto.OrderType_REQ:
		phi.requestMgr.propose(msg)
	case commonProto.OrderType_LOG:
		phi.logMgr.propose(msg)
	}
}
