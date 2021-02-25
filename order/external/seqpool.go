package external

import (
	commonBasic "github.com/Grivn/phalanx/common/basic"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type SequencePool interface {
	commonBasic.Basic

	MsgRecorder(msg *commonProto.OrderedMsg)
}
