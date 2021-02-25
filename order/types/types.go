package types

import (
	commonProto "github.com/Grivn/phalanx-common/types/protos"
	"github.com/Grivn/phalanx-order/external"

	logger "github.com/ultramesh/fancylogger"
)

type Config struct {
	Author uint64
	BatchSize int

	TxChan chan *commonProto.Transaction
	ReqChan chan *commonProto.OrderedReq
	LogChan chan *commonProto.OrderedLog

	EnclaveFile string
	SealedKey []byte

	Network      external.Network
	TEE          external.USIG
	SequencePool external.SequencePool

	Logger *logger.Logger
}
