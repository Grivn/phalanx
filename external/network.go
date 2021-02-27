package external

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type Network interface {
	Broadcast(msg *commonProto.CommMsg)
}
