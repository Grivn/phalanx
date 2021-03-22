package loggenerator

import (
	"time"

	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/loggenerator/types"
)

type generator struct {
	author   uint64
	sequence uint64
	replyC   chan types.ReplyEvent
	sender   *senderProxy
	recorder *recorder
	logger   external.Logger
}

func newGenerator(author uint64, replyC chan types.ReplyEvent, recorder *recorder, network external.Network, logger external.Logger) *generator {
	return &generator{
		author:   author,
		sequence: uint64(0),
		replyC:   replyC,
		sender:   newSenderProxy(author, network),
		recorder: recorder,
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

	generator.logger.Infof("[GENERATE LOG] replica %d sequence %d", generator.author, generator.sequence)
	generator.sender.broadcast(msg)
	return msg
}
