package phalanx

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	lm "github.com/Grivn/phalanx/logmgr/types"
	rm "github.com/Grivn/phalanx/reqmgr/types"
	tp "github.com/Grivn/phalanx/txpool/types"
)

type phalanxImpl struct {
	author uint64

	txpool  api.TxPool
	requestMgr api.RequestManager
	logMgr api.LogManager

	recvC chan interface{}
	tpC chan tp.ReplyEvent
	rmC chan rm.ReplyEvent
	lmC chan lm.ReplyEvent
	closeC   chan bool

	logger external.Logger
}

func newPhalanxImpl() *phalanxImpl {
	return &phalanxImpl{}
}

func (phi *phalanxImpl) listenTxPool() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist tx pool listener for phalanx")
			return
		case ev := <-phi.tpC:
			phi.dispatchTxPoolEvent(ev)
		}
	}
}

func (phi *phalanxImpl) listenReqManager() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist request manager listener for phalanx")
			return
		case ev := <-phi.rmC:
			phi.dispatchRequestEvent(ev)
		}
	}
}

func (phi *phalanxImpl) listenLogManager() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist log manager listener for phalanx")
			return
		case ev := <-phi.lmC:
			phi.dispatchLogEvent(ev)
		}
	}
}

func (phi *phalanxImpl) dispatchTxPoolEvent(event tp.ReplyEvent) {
	switch event.EventType {
	case tp.ReplyGenerateBatchEvent:
		batch := event.Event.(*commonProto.Batch)
		phi.requestMgr.Generate(batch.BatchId)
	case tp.ReplyLoadBatchEvent:
	case tp.ReplyMissingBatchEvent:
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchRequestEvent(event rm.ReplyEvent) {
	switch event.EventType {
	case rm.ReqReplyBatchByOrder:
		bid := event.Event.(*commonProto.BatchId)
		phi.logMgr.Generate(bid)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchLogEvent(event lm.ReplyEvent) {
	switch event.EventType {
	case lm.LogReplyQuorumBinaryEvent:
	case lm.LogReplyExecuteEvent:
	default:
		return
	}
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
