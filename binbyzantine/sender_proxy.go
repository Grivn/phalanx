package binbyzantine

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/gogo/protobuf/proto"
)

type senderProxy struct {
	author  uint64
	network external.Network
	logger  external.Logger
}

func newSenderProxy(author uint64, network external.Network, logger external.Logger) *senderProxy {
	return &senderProxy{
		author:  author,
		network: network,
		logger:  logger,
	}
}

func (sender *senderProxy) broadcast(ntf *commonProto.BinaryNotification) {
	payload, err := proto.Marshal(ntf)
	if err != nil {
		sender.logger.Errorf("Marshal error: %s", err)
		return
	}
	comm := &commonProto.CommMsg{
		Author:  sender.author,
		Type:    commonProto.CommType_BBA,
		Payload: payload,
	}
	sender.network.Broadcast(comm)
}
