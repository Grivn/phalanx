package txpool

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type sender struct {
	author  uint64
	network external.Network
}

func newSender(author uint64, network external.Network) *sender {
	return &sender{
		author:  author,
		network: network,
	}
}

func (s *sender) broadcast(batch *commonProto.Batch) {
	payload, err := proto.Marshal(batch)
	if err != nil {
		panic(err)
	}

	comm := &commonProto.CommMsg{
		Author:  s.author,
		Type:    commonProto.CommType_BATCH,
		Payload: payload,
	}
	s.network.Broadcast(comm)
}
