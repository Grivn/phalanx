package order

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"

	"github.com/gogo/protobuf/proto"
)

type usigProxy struct {
	author uint64
	usig internal.USIG

	logger external.Logger
}

func (up *usigProxy) generateUIMsg(msg *commonProto.OrderedMsg) *commonProto.UIMsg {
	payload, err := proto.Marshal(msg)
	if err != nil {
		up.logger.Error("Marshal ordered_msg error: %s", err)
		return nil
	}

	UI, err := up.usig.CreateUI(payload)
	if err != nil {
		up.logger.Error("Create unique identifier error: %s", err)
		return nil
	}

	return &commonProto.UIMsg{
		Author:  up.author,
		Payload: payload,
		UI:      UI,
	}
}

func (up *usigProxy) verifyUIMsg(msg *commonProto.UIMsg) {

}
