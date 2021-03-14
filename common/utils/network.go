package utils

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

type replyNetwork struct {
	replyC chan interface{}
	logger external.Logger
}

func NewReplyNetwork(replyC chan interface{}) external.Network {
	logger := NewRawLogger()
	return &replyNetwork{
		replyC: replyC,
		logger: logger,
	}
}

func (network *replyNetwork) Broadcast(msg *commonProto.CommMsg) {
	network.logger.Debugf("broadcast message, type %d", msg.Type)
	go func() {
		network.replyC <- msg
	}()
}
