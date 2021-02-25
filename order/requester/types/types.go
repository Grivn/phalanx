package types

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/order/external"

	logger "github.com/ultramesh/fancylogger"
)

type Config struct {
	Author    uint64
	BatchSize int
	TxC       chan *commonProto.Transaction
	ReqC      chan *commonProto.OrderedReq

	Sender external.Network

	TEE external.USIG

	SequencePool external.SequencePool

	Logger   *logger.Logger
}
