package utils

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type network struct {
	value *commonProto.CommMsg
}

func NewFakeNetwork() external.Network {
	return &network{}
}

func (network *network) Broadcast(msg *commonProto.CommMsg) {
	network.value = msg
}
