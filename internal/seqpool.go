package internal

import (
	commonBasic "github.com/Grivn/phalanx/common/basic"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type SequencePool interface {
	commonBasic.Basic

	RecordMsg(msg *commonProto.OrderedMsg)
}
