package logmgr

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/logmgr/types"
)

type logMgrImpl struct {
	n int

	author uint64

	recvC chan types.RecvEvent

	replyC chan types.ReplyEvent

	closeC chan bool

	generator *generator

	// author to recorder struct
	recorder map[uint64]*recorder

	// sequence to binary struct
	binary map[uint64]*binary

	auth api.Authenticator

	logger external.Logger
}

func newLogMgrImpl(n int, author uint64, replyC chan types.ReplyEvent, auth api.Authenticator, network external.Network, logger external.Logger) *logMgrImpl {
	re := make(map[uint64]*recorder)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		re[id] = newRecorder(id, auth, logger)
	}
	return &logMgrImpl{
		n:         n,
		author:    author,
		recvC:     make(chan types.RecvEvent),
		replyC:    replyC,
		closeC:    make(chan bool),
		generator: newGenerator(author, replyC, re[author], network, auth, logger),
		recorder:  re,
		binary:    make(map[uint64]*binary),
		auth:      auth,
		logger:    logger,
	}
}

func (lm *logMgrImpl) generate(bid *commonProto.BatchId) {
	event := types.RecvEvent{
		EventType: types.LogRecvGenerate,
		Event:     bid,
	}
	lm.recvC <- event
}

func (lm *logMgrImpl) record(msg *commonProto.SignedMsg) {
	event := types.RecvEvent{
		EventType: types.LogRecvRecord,
		Event:     msg,
	}
	lm.recvC <- event
}

func (lm *logMgrImpl) ready(binarySet []byte) {
	event := types.RecvEvent{
		EventType: types.LogRecvReady,
		Event:     binarySet,
	}
	lm.recvC <- event
}

func (lm *logMgrImpl) start() {
	go lm.listener()
}

func (lm *logMgrImpl) stop() {
	select {
	case <-lm.closeC:
	default:
		close(lm.closeC)
	}
}

func (lm *logMgrImpl) listener() {
	for {
		select {
		case <-lm.closeC:
			lm.logger.Notice("exist log manager listener")
			return
		case ev := <-lm.recvC:
			lm.dispatchRecvEvent(ev)
		}
	}
}

func (lm *logMgrImpl) dispatchRecvEvent(event types.RecvEvent) {
	switch event.EventType {
	case types.LogRecvGenerate:
		bid, ok := event.Event.(*commonProto.BatchId)
		if !ok {
			return
		}
		msg := lm.generator.generate(bid)
		lm.binary[msg.Sequence].update(msg)
	case types.LogRecvRecord:
		signed, ok := event.Event.(*commonProto.SignedMsg)
		if !ok {
			return
		}
		msg := lm.recorder[signed.Author].update(signed)
		if msg == nil {
			return
		}
		bin, ok := lm.binary[msg.Sequence]
		if !ok {
			bin = newBinary(lm.n, lm.author, msg.Sequence, lm.replyC, lm.logger)
			lm.binary[msg.Sequence] = bin
		}
		bin.update(msg)
	case types.LogRecvReady:
		bTag := event.Event.(*commonProto.BinaryTag)
		var logs []*commonProto.OrderedMsg
		for index, value := range bTag.BinarySet {
			id := uint64(index+1)
			if value == 1 {
				log := lm.recorder[id].logs[bTag.LogId]
				logs = append(logs, log)
			}
			lm.recorder[id].upgrade(bTag.LogId)
		}
		event := types.ReplyEvent{
			EventType: types.LogReplyExecuteEvent,
			Event:     logs,
		}
		lm.replyC <- event
	default:
		return
	}
}
