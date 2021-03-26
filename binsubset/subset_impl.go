package binsubset

import (
	"github.com/Grivn/phalanx/binsubset/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type subsetImpl struct {
	n int
	f int

	author uint64

	lock map[uint64]bool

	recorder map[uint64]*recorderMgr

	quorumTag map[uint64]*commonProto.BinaryTag

	recvC chan types.RecvEvent

	closeC chan bool

	replyC chan types.ReplyEvent

	sender *senderProxy

	logger external.Logger
}

type ready struct {
	seq  uint64
	hash string
}

func newSubsetImpl(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) *subsetImpl {
	logger.Noticef("replica %d init the binary byzantine module", author)

	return &subsetImpl{
		n: n,
		f: (n-1)/3,
		author:   author,

		lock: make(map[uint64]bool),

		recorder: make(map[uint64]*recorderMgr),

		quorumTag: make(map[uint64]*commonProto.BinaryTag),

		recvC:    make(chan types.RecvEvent),
		closeC:   make(chan bool),
		replyC:   replyC,
		sender:   newSenderProxy(author, network, logger),
		logger:   logger,
	}
}

func (si *subsetImpl) trigger(tag *commonProto.BinaryTag) {
	event := types.RecvEvent{
		EventType: types.BinaryRecvTag,
		Event:     tag,
	}

	si.recvC <- event
}

func (si *subsetImpl) propose(ntf *commonProto.BinaryNotification) {
	event := types.RecvEvent{
		EventType: types.BinaryRecvNotification,
		Event:     ntf,
	}

	si.recvC <- event
}

func (si *subsetImpl) start() {
	go si.listener()
}

func (si *subsetImpl) stop() {
	select {
	case <-si.closeC:
	default:
		close(si.closeC)
	}
}

func (si *subsetImpl) listener() {
	for {
		select {
		case <-si.closeC:
			si.logger.Notice("exist binary byzantine listener")
			return
		case ev, ok := <-si.recvC:
			if !ok {
				continue
			}
			si.dispatchEvent(ev)
		}
	}
}

func (si *subsetImpl) dispatchEvent(event types.RecvEvent) {
	switch event.EventType {
	case types.BinaryRecvTag:
		tag, ok := event.Event.(*commonProto.BinaryTag)
		if !ok {
			si.logger.Error("parsing error")
			return
		}
		si.processLocal(tag)
	case types.BinaryRecvNotification:
		ntf, ok := event.Event.(*commonProto.BinaryNotification)
		if !ok {
			si.logger.Error("parsing error")
			return
		}
		si.processRemote(ntf)
	default:
		si.logger.Errorf("Invalid event type: code %d", event.EventType)
		return
	}
}

func (si *subsetImpl) processLocal(tag *commonProto.BinaryTag) {
	if si.lock[tag.Sequence] {
		si.logger.Warningf("replica %d is processing binary tag for sequence %d", si.author, tag.Sequence)
		return
	}
	si.lock[tag.Sequence] = true

	ntf := &commonProto.BinaryNotification{
		Type:      commonProto.BinaryType_TAG,
		Author:    si.author,
		BinaryTag: tag,
	}
	si.sender.broadcast(ntf)

	si.processNotificationTag(ntf)
}

func (si *subsetImpl) processRemote(ntf *commonProto.BinaryNotification) {
	switch ntf.Type {
	case commonProto.BinaryType_TAG:
		si.processNotificationTag(ntf)
	case commonProto.BinaryType_QC:
		si.processQuorumTag(ntf.BinaryTag)
	case commonProto.BinaryType_READY:
		// todo skip
	default:
		return
	}
}

func (si *subsetImpl) processNotificationTag(msg *commonProto.BinaryNotification) {
	si.logger.Debugf("replica %d received notification for sequence %d from replica %d", si.author, msg.BinaryTag.Sequence, msg.Author)

	tag := msg.BinaryTag
	qTag := si.getRecorder(tag.Sequence).record(msg.Author, tag)

	//if si.continueQuorumTag(tag.Sequence) {
	//	return
	//}

	if qTag != nil {
		ntf := &commonProto.BinaryNotification{
			Author:    si.author,
			Type:      commonProto.BinaryType_QC,
			BinaryTag: qTag,
		}
		si.sender.broadcast(ntf)
		si.processQuorumTag(ntf.BinaryTag)
	}
}

func (si *subsetImpl) processQuorumTag(qTag *commonProto.BinaryTag) {
	si.logger.Infof("replica %d received quorum event for sequence %d", si.author, qTag.Sequence)
	si.quorumTag[qTag.Sequence] = qTag
	//ntf := si.tryToSendBinaryReady(qTag)
	//
	//if ntf != nil {
	delete(si.quorumTag, qTag.Sequence)
	event := types.ReplyEvent{
		EventType: types.BinaryReplyReady,
		Event:     qTag,
	}
	go func() {
		si.replyC <- event
	}()
	//}
}

func (si *subsetImpl) processBinaryReady(ready *commonProto.BinaryNotification) {

}

func (si *subsetImpl) continueQuorumTag(sequence uint64) bool {
	qTag, ok := si.quorumTag[sequence]
	if ok {
		si.logger.Debugf("replica %d has already received quorum event, continue to process", si.author)
		ntf := si.tryToSendBinaryReady(qTag)

		if ntf != nil {
			delete(si.quorumTag, sequence)
			event := types.ReplyEvent{
				EventType: types.BinaryReplyReady,
				Event:     ntf.BinaryTag,
			}
			si.replyC <- event
		}
	}
	return ok
}

func (si *subsetImpl) tryToSendBinaryReady(qTag *commonProto.BinaryTag) *commonProto.BinaryNotification {
	re := si.getRecorder(qTag.Sequence)
	if re.compare(qTag.BinarySet) {
		ntf := &commonProto.BinaryNotification{
			Author:    re.author,
			Type:      commonProto.BinaryType_READY,
			BinaryTag: qTag,
		}
		si.sender.broadcast(ntf)
		return ntf
	}
	return nil
}

func (si *subsetImpl) getRecorder(sequence uint64) *recorderMgr {
	re, ok := si.recorder[sequence]
	if !ok {
		re = newRecorder(si.n, si.f, si.author, sequence, si.logger)
		si.recorder[sequence] = re
	}
	return re
}
