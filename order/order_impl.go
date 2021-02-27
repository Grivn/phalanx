package order

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/order/types"
)

type orderImpl struct {
	author uint64

	usig   internal.USIG
	txpool internal.TxPool

	recvC  chan *commonProto.OrderedMsg
	close   chan bool

	logger external.Logger
}

func newOrderImpl(config types.Config) *orderImpl {
	return &orderImpl{}
}

func (oi *orderImpl) start() {}

func (oi *orderImpl) stop() {}

func (oi *orderImpl) receiveTransaction(tx *commonProto.Transaction) {}

func (oi *orderImpl) receiveOrderedMsg(req *commonProto.OrderedMsg) {}
