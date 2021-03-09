package consensus

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/consensus/types"
	"github.com/Grivn/phalanx/external"
)

type orderImpl struct {
	author uint64

	usig   *usigProxy
	txpool api.TxPool

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
