package binbyzantine

import (
	"github.com/Grivn/phalanx/binbyzantine/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type byzantineImpl struct {
	author uint64

	recvC chan types.RecvEvent

	closeC chan bool

	verifier *verificationMgr

	sender *senderProxy

	logger external.Logger
}

func newByzantineImpl(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) *byzantineImpl {
	logger.Noticef("replica %d init the binary byzantine module", author)

	return &byzantineImpl{
		author:   author,
		recvC:    make(chan types.RecvEvent),
		closeC:   make(chan bool),
		verifier: newVerificationMgr(n, author, replyC, network, logger),
		sender:   newSenderProxy(author, network, logger),
		logger:   logger,
	}
}

func (bi *byzantineImpl) trigger(tag *commonProto.BinaryTag) {
	event := types.RecvEvent{
		EventType: types.BinaryRecvTag,
		Event:     tag,
	}

	bi.recvC <- event
}

func (bi *byzantineImpl) propose(ntf *commonProto.BinaryNotification) {
	event := types.RecvEvent{
		EventType: types.BinaryRecvNotification,
		Event:     ntf,
	}

	bi.recvC <- event
}

func (bi *byzantineImpl) start() {
	go bi.listener()
}

func (bi *byzantineImpl) stop() {
	select {
	case <-bi.closeC:
	default:
		close(bi.closeC)
	}
}

func (bi *byzantineImpl) listener() {
	for {
		select {
		case <-bi.closeC:
			bi.logger.Notice("exist binary byzantine listener")
			return
		case ev := <-bi.recvC:
			bi.dispatchEvent(ev)
		}
	}
}

func (bi *byzantineImpl) dispatchEvent(event types.RecvEvent) {
	switch event.EventType {
	case types.BinaryRecvTag:
		tag, ok := event.Event.(*commonProto.BinaryTag)
		if !ok {
			bi.logger.Error("parsing error")
			return
		}
		bi.verifier.processLocal(tag)
	case types.BinaryRecvNotification:
		ntf, ok := event.Event.(*commonProto.BinaryNotification)
		if !ok {
			bi.logger.Error("parsing error")
			return
		}
		bi.verifier.processRemote(ntf)
	default:
		bi.logger.Errorf("Invalid event type: code %d", event.EventType)
		return
	}
}
