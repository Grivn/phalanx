package types

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type Config struct {
	N int

	Author uint64

	ReplyC chan *commonProto.BatchId

	Logger external.Logger
}
