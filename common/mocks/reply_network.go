package mocks

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"math/rand"
	"time"
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
	//network.logger.Debugf("broadcast message, type %d", msg.Type)
	go func() {
		salt := rand.Int()%5000
		time.Sleep(time.Duration(salt)*time.Millisecond)
		network.replyC <- msg
	}()
}
