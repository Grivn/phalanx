package mocks

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type network struct {
	value *commonProto.CommMsg
	logger external.Logger
}

func NewFakeNetwork() external.Network {
	logger := NewRawLogger()
	return &network{
		logger: logger,
	}
}

func (network *network) Broadcast(msg *commonProto.CommMsg) {
	//network.logger.Debugf("broadcast message, type %d", msg.Type)
	network.value = msg
}
