package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type RequestManager interface {
	Basic

	Record(msg *commonProto.OrderedMsg)
}
