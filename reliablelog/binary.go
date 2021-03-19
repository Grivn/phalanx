package reliablelog

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/reliablelog/types"
)

type binary struct {
	active   bool
	readyTag *commonProto.BinaryTag
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
		active:   false,
		readyTag: nil,
		n:        n,
		f:        (n-1)/4,
		author:   author,
		sequence: sequence,
		occurred: make(map[uint64]bool),
		replyC:   replyC,
		logger:   logger,
	}
}

// we will continuously generate binary tag, even thought we have already reached quorum size before,
// until current log sequence has been active
func (binary *binary) update(msg *commonProto.OrderedMsg) {
	binary.logger.Infof("replica %d receive log from replica %d, try to update binary set for sequence %d", binary.author, msg.Author, binary.sequence)

	if msg.Sequence != binary.sequence {
		binary.logger.Warningf("replica %d received mismatched sequence number expected %d recv %d", binary.author, binary.sequence, msg.Sequence)
		return
	}

	binary.occurred[msg.Author] = true

	if !binary.occurred[binary.author]{
		binary.logger.Infof("replica %d hasn't generated a log on sequence %d", binary.author, binary.sequence)
		return
	}

	if len(binary.occurred) >= binary.quorum()  {
		tag := binary.convert()
		event := types.ReplyEvent{
			EventType: types.LogReplyQuorumBinaryEvent,
			Event:     tag,
		}
		binary.active = true
		binary.logger.Infof("replica %d reach quorum size, generate quorum binary tag %v for sequence %d", binary.author, binary.sequence, tag.BinarySet)
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
		Sequence:   binary.sequence,
		BinaryHash: hash,
		BinarySet:  set,
	}
	return bTag
}

func (binary *binary) ready(tag *commonProto.BinaryTag) {
	binary.readyTag = tag
}

func (binary *binary) getTag() *commonProto.BinaryTag {
	return binary.readyTag
}

func (binary *binary) quorum() int {
	return binary.n - binary.f
}

func (binary *binary) onActive() {
	binary.logger.Infof("replica %d starts the binary processor for sequence %d", binary.author, binary.sequence)
	binary.active = true
}

func (binary *binary) isActive() bool {
	return binary.active
}
