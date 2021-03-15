package mocks

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

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
