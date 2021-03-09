package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type SequencePool interface {
	Basic

	Record(msg *commonProto.OrderedMsg)
}
