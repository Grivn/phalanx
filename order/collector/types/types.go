package types

import (
	commonProto "github.com/Grivn/phalanx-common/types/protos"
	"github.com/Grivn/phalanx-order/external"

	logger "github.com/ultramesh/fancylogger"
)

type Config struct {
	Author uint64

	ReqC chan *commonProto.OrderedReq
	LogC chan *commonProto.OrderedLog

	Sender       external.Network
	TEE          external.TEE
	SequencePool external.SequencePool

	Logger *logger.Logger
}
