package phalanx

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	external "github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/order"

	logger "github.com/ultramesh/fancylogger"
)

type phalanxImpl struct {
	id uint64

	order order.Order

	txpool external.TxPool

	txChan chan *commonProto.Transaction

	logger *logger.Logger
}

func (ph *phalanxImpl) listener() {
	for {
		select {
		case tx := <-ph.txChan:
			ph.processTx(tx)
		case req := <-ph.reqChan:
			ph.processOrderedReq(req)
		case log := <-ph.logChan:
			ph.processOrderedLog(log)
		}
	}
}

// process the transactions send from api
func (ph *phalanxImpl) processTx(tx *commonProto.Transaction) {
	ph.txpool.PostTx(tx)
}

func (ph *phalanxImpl) processOrderedReq(req *commonProto.OrderedReq) {
	ph.order.ReceiveOrderedReq(req)
}

func (ph *phalanxImpl) processOrderedLog(log *commonProto.OrderedLog) {
	ph.order.ReceiveOrderedLog(log)
}
