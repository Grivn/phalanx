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

	lastSeq uint64

	store map[uint64]*binary

	local map[uint64]*binary

	certs map[uint64]*cert

	qcCerts map[uint64]*qcCert

	replyC chan types.ReplyEvent

	sender *senderProxy

	logger external.Logger
}

func newVerificationMgr(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) *verificationMgr {
	return &verificationMgr{
		n: n,

		f: (n-1)/3,

		author: author,

		lastSeq: uint64(0),

		store: make(map[uint64]*binary),

		local: make(map[uint64]*binary),

		certs: make(map[uint64]*cert),

		qcCerts: make(map[uint64]*qcCert),

		replyC: replyC,

		sender: newSenderProxy(author, network, logger),

		logger: logger,
	}
}

func (v *verificationMgr) processLocal(tag *commonProto.BinaryTag) {
	if !v.binaryStore(tag) {
		v.logger.Debugf("replica %d is processing current tag sequence %d", v.author, tag.Sequence)
		return
	}

	bin := v.getBinary(tag)

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
	v.dispatchNotification(ntf)
}

func (v *verificationMgr) dispatchNotification(ntf *commonProto.BinaryNotification) {
	switch ntf.Type {
	case commonProto.BinaryType_TAG:
		set := v.updateCert(ntf.Author, ntf.BinaryTag)
		bin := v.getBinary(ntf.BinaryTag)
		if bin.compare(set) {
			msg := &commonProto.BinaryNotification{
				Author:    v.author,
				Type:      commonProto.BinaryType_TAG,
				BinaryTag: bin.convert(),
			}
			v.sender.broadcast(msg)
		}
	case commonProto.BinaryType_QC:
		v.updateQCCert(ntf.Author, ntf.BinaryTag)
	default:
		return
	}
}

func (v *verificationMgr) updateCert(author uint64, tag *commonProto.BinaryTag) []byte {
	c := v.getCert(tag.Sequence)

	if c.finished {
		v.logger.Infof("replica %d has finished tag-certificate for sequence %d", v.author, tag.Sequence)
		return nil
	}

	counter, ok := c.counter[tag.BinaryHash]
	if !ok {
		counter = make(map[uint64]bool)
		c.counter[tag.BinaryHash] = counter
	}
	counter[author] = true

	v.logger.Infof("replica %d received set %v for sequence %d, count %d", v.author, tag.BinarySet, tag.Sequence, len(c.counter[tag.BinaryHash]))

	// we need to check if there are quorum replicas have agreed on current binary tag, if so, directly return tag
	if len(c.counter[tag.BinaryHash]) >= v.quorum() && v.local[tag.Sequence].include(tag) {
		c.finished = true
		v.logger.Infof("replica %d find quorum set %v for sequence %d, broadcast quorum cert event", v.author, tag.BinarySet, tag.Sequence)

		ntf := &commonProto.BinaryNotification{
			Author:    v.author,
			Type:      commonProto.BinaryType_QC,
			BinaryTag: tag,
		}
		v.sender.broadcast(ntf)
		v.updateQCCert(v.author, tag)

		v.getQCCert(tag.Sequence).broadcast = true

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

func (v *verificationMgr) updateQCCert(author uint64, tag *commonProto.BinaryTag) bool {
	qc := v.getQCCert(tag.Sequence)

	if qc.finished {
		v.logger.Infof("replica %d reject quorum cert event for sequence %d", v.author, tag.Sequence)
		return true
	}

	counter, ok := qc.counter[tag.BinaryHash]
	if !ok {
		counter = make(map[uint64]bool)
		qc.counter[tag.BinaryHash] = counter
	}
	counter[author] = true

	v.logger.Infof("replica %d received quorum set %v for sequence %d, count %d", v.author, tag.BinarySet, tag.Sequence, len(qc.counter[tag.BinaryHash]))

	if len(qc.counter[tag.BinaryHash]) >= v.oneQuorum() {
		v.getCert(tag.Sequence).finished = true
		ntf := &commonProto.BinaryNotification{
			Author:    v.author,
			Type:      commonProto.BinaryType_QC,
			BinaryTag: tag,
		}
		v.sender.broadcast(ntf)
		qc.broadcast = true
	}

	// we need to check if there are quorum replicas have agreed on current binary tag, if so, directly return tag
	if len(qc.counter[tag.BinaryHash]) >= v.quorum() && qc.broadcast {
		qc.finished = true
		v.logger.Infof("replica %d find quorum quorum-set %v for sequence %d, broadcast quorum cert event", v.author, tag.BinarySet, tag.Sequence)

		event := types.ReplyEvent{
			EventType: types.BinaryReplyReady,
			Event:     tag,
		}
		v.replyC <- event
		return true
	}

	// record the tag which hasn't reached quorum-set
	qc.tags[tag.BinaryHash] = tag

	return false
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

func (v *verificationMgr) getQCCert(sequence uint64) *qcCert {
	c, ok := v.qcCerts[sequence]
	if !ok {
		c = &qcCert{
			counter: make(map[string]map[uint64]bool),
			tags:    make(map[string]*commonProto.BinaryTag),
		}
		v.qcCerts[sequence] = c
	}
	return c
}

func (v *verificationMgr) binaryStore(tag *commonProto.BinaryTag) bool {
	bin, ok := v.store[tag.Sequence]
	if !ok {
		bin = newBinary(v.n, v.author, tag, v.logger)
		v.store[tag.Sequence] = bin
		// it is the first time current node process the tag
		return true
	}
	return false
}

func (v *verificationMgr) getBinary(tag *commonProto.BinaryTag) *binary {
	bin, ok := v.local[tag.Sequence]
	if !ok {
		bin = newBinary(v.n, v.author, tag, v.logger)
		v.local[tag.Sequence] = bin
	}
	return bin
}

func (v *verificationMgr) quorum() int {
	return v.n - v.f
}

func (v *verificationMgr) oneQuorum() int {
	return v.f + 1
}
