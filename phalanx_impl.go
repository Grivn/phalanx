package phalanx

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/order"

	logger "github.com/ultramesh/fancylogger"
)

type phalanxImpl struct {
	id uint64

	order order.Order

	txsChan chan *commonProto.TransactionSet
	reqChan chan *commonProto.OrderedReq
	logChan chan *commonProto.OrderedLog

	logger *logger.Logger
}

func (ph *phalanxImpl) listener() {
	for {
		select {
		case txs := <-ph.txsChan:
			ph.processTransaction(txs)
		case req := <-ph.reqChan:
			ph.processOrderedReq(req)
		case log := <-ph.logChan:
			ph.processOrderedLog(log)
		}
	}
}

// process the transactions send from api
func (ph *phalanxImpl) processTransaction(txs *commonProto.TransactionSet) {
	ph.order.ReceiveTransaction(txs)
}

func (ph *phalanxImpl) processOrderedReq(req *commonProto.OrderedReq) {
	ph.order.ReceiveOrderedReq(req)
}

func (ph *phalanxImpl) processOrderedLog(log *commonProto.OrderedLog) {
	ph.order.ReceiveOrderedLog(log)
}
