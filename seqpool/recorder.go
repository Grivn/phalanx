package seqpool

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type RequestManager interface {
	api.Basic

	Save(msg *commonProto.OrderedMsg)
}

type LogManager interface {
	Save(msg *commonProto.OrderedMsg)

	Load(author uint64, hash string) uint64
}
