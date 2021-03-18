package reliablelog

import (
	"time"

	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/reliablelog/types"

	"github.com/gogo/protobuf/proto"
)

type generator struct {
	author   uint64
	sequence uint64
	replyC   chan types.ReplyEvent
	sender   *senderProxy
	recorder *recorder
	auth     api.Authenticator
	logger   external.Logger
}

func newGenerator(author uint64, replyC chan types.ReplyEvent, recorder *recorder, network external.Network, auth api.Authenticator, logger external.Logger) *generator {
	return &generator{
		author:   author,
		sequence: uint64(0),
		replyC:   replyC,
		sender:   newSenderProxy(author, network),
		recorder: recorder,
		auth:     auth,
		logger:   logger,
	}
}

func (generator *generator) generate(bid *commonProto.BatchId) *commonProto.OrderedMsg {
	generator.sequence++

	msg := &commonProto.OrderedMsg{
		Type:      commonProto.OrderType_LOG,
		Author:    generator.author,
		Sequence:  generator.sequence,
		BatchId:   bid,
		Timestamp: time.Now().UnixNano(),
	}

	signed := generator.sign(msg)

	generator.logger.Infof("[GENERATE LOG] replica %d sequence %d", generator.author, generator.sequence)
	generator.sender.broadcast(signed)
	return msg
}

func (generator *generator) sign(msg *commonProto.OrderedMsg) *commonProto.SignedMsg {
	payload, err := proto.Marshal(msg)
	if err != nil {
		generator.logger.Errorf("Marshal error, %s", err)
		return nil
	}

	sig, err := generator.auth.GenerateMessageAuthenTag(api.USIGAuthen, payload)
	if err != nil {
		generator.logger.Errorf("Generate usig error, %s", err)
		return nil
	}

	signed := &commonProto.SignedMsg{
		Author:    generator.author,
		Payload:   payload,
		Signature: sig,
	}
	return signed
}

