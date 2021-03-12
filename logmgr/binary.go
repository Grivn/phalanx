package logmgr

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/logmgr/types"
)

type binary struct {
	n        int
	f        int
	author   uint64
	sequence uint64
	occurred map[uint64]bool
	replyC   chan types.ReplyEvent
	logger   external.Logger
}

func newBinary(n int, author uint64, sequence uint64, replyC chan types.ReplyEvent, logger external.Logger) *binary {
	return &binary{
		n:        n,
		f:        (n-1)/4,
		author:   author,
		sequence: sequence,
		occurred: make(map[uint64]bool),
		replyC:   replyC,
		logger:   logger,
	}
}

func (binary *binary) update(msg *commonProto.OrderedMsg) {
	binary.logger.Infof("receive log from replica %d, try to update binary set for sequence %d", msg.Author, binary.sequence)

	if msg.Sequence != binary.sequence {
		binary.logger.Warningf("mismatched sequence number expected %d recv %d", binary.sequence, msg.Sequence)
		return
	}

	binary.occurred[msg.Author] = true

	if !binary.occurred[binary.author]{
		binary.logger.Infof("Current replica hasn't generated a log on sequence %d", binary.sequence)
		return
	}

	if len(binary.occurred) >= binary.quorum()  {
		tag := binary.convert()
		event := types.ReplyEvent{
			EventType: types.LogReplyQuorumBinaryEvent,
			Event:     tag,
		}
		binary.logger.Infof("reach quorum size for sequence %d, generate quorum binary tag, set %v", binary.sequence, tag.BinarySet)
		binary.replyC <- event
	}
}

func (binary *binary) convert() *commonProto.BinaryTag {
	set := make([]byte, binary.n)
	for index := range set {
		if binary.occurred[uint64(index+1)] {
			set[index] = 1
		}
	}
	hash := commonTypes.CalculatePayloadHash(set, 0)

	bTag := &commonProto.BinaryTag{
		LogId:      binary.sequence,
		BinaryHash: hash,
		BinarySet:  set,
	}
	return bTag
}

func (binary *binary) quorum() int {
	return binary.n - binary.f
}
