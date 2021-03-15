package reliablelog

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type senderProxy struct {
	author  uint64
	network external.Network
}

func newSenderProxy(author uint64, network external.Network) *senderProxy {
	return &senderProxy{
		author:  author,
		network: network,
	}
}

func (sp *senderProxy) broadcast(signed *commonProto.SignedMsg) {
	payload, err := proto.Marshal(signed)
	if err != nil {
		panic(err)
	}

	comm := &commonProto.CommMsg{
		Author:  sp.author,
		Type:    commonProto.CommType_ORDER,
		Payload: payload,
	}
	sp.network.Broadcast(comm)
}
