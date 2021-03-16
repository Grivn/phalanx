package binbyzantine

import (
	"github.com/Grivn/phalanx/binbyzantine/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type verificationMgr struct {
	n int

	f int

	author uint64

	local map[uint64]*binary

	certs map[uint64]*cert

	replyC chan types.ReplyEvent

	sender *senderProxy

	logger external.Logger
}

func newVerificationMgr(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) *verificationMgr {
	return &verificationMgr{
		n: n,

		f: (n-1)/4,

		author: author,

		local: make(map[uint64]*binary),

		certs: make(map[uint64]*cert),

		replyC: replyC,

		sender: newSenderProxy(author, network, logger),

		logger: logger,
	}
}

func (v *verificationMgr) processLocal(tag *commonProto.BinaryTag) {
	bin := v.initBinary(tag)

	// update the certs of current sequence number
	set := v.updateCert(v.author, tag)

	// compare the local binary set and the tags of this sequence number
	bin.compare(set)

	ntf := &commonProto.BinaryNotification{
		Author:    v.author,
		BinaryTag: bin.convert(),
	}
	v.sender.broadcast(ntf)
}

func (v *verificationMgr) processRemote(ntf *commonProto.BinaryNotification) {
	set := v.updateCert(ntf.Author, ntf.BinaryTag)

	bin := v.getBinary(ntf.BinaryTag.Sequence)
	if bin == nil {
		v.logger.Infof("replica %d hasn't init binary for sequence %d, ignore it", v.author, ntf.BinaryTag.Sequence)
		return
	}

	if bin.compare(set) {
		ntf := &commonProto.BinaryNotification{
			Author:    v.author,
			BinaryTag: bin.convert(),
		}
		v.sender.broadcast(ntf)
	}
}

func (v *verificationMgr) updateCert(author uint64, tag *commonProto.BinaryTag) []byte {
	c := v.getCert(tag.Sequence)

	counter, ok := c.counter[tag.BinaryHash]
	if !ok {
		counter = make(map[uint64]bool)
		c.counter[tag.BinaryHash] = counter
	}
	counter[author] = true

	// we need to check if there are quorum replicas have agreed on current binary tag, if so, directly return tag
	if len(c.counter[tag.BinaryHash]) >= v.quorum() {
		v.logger.Infof("replica %d find quorum set %v for sequence %d", v.author, tag.BinarySet, tag.Sequence)

		event := types.ReplyEvent{
			EventType: types.BinaryReplyReady,
			Event:     tag,
		}

		v.replyC <- event
		return tag.BinarySet
	}

	// record the tag which hasn't reached quorum-set
	c.tags[tag.BinaryHash] = tag

	// find the replica id which has been verified more than f+1 times
	set := make([]byte, v.n)
	for index, value := range tag.BinarySet {
		if value == 1 {
			id := uint64(index+1)
			c.bits[id]++

			if len(c.bits) >= v.oneQuorum() {
				set[index] = 1
			}
		}
	}

	return set
}

func (v *verificationMgr) getCert(sequence uint64) *cert {
	c, ok := v.certs[sequence]
	if !ok {
		c = &cert{
			counter: make(map[string]map[uint64]bool),
			tags:    make(map[string]*commonProto.BinaryTag),
			bits:    make(map[uint64]uint64),
		}
		v.certs[sequence] = c
	}
	return c
}

func (v *verificationMgr) initBinary(tag *commonProto.BinaryTag) *binary {
	bin := newBinary(v.n, v.author, tag, v.logger)
	v.local[tag.Sequence] = bin
	return bin
}

func (v *verificationMgr) getBinary(sequence uint64) *binary {
	return v.local[sequence]
}

func (v *verificationMgr) quorum() int {
	return v.n - v.f
}

func (v *verificationMgr) oneQuorum() int {
	return v.f + 1
}
