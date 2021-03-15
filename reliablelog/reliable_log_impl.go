package reliablelog

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/reliablelog/types"
)

type reliableLogImpl struct {
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

func newReliableLogImpl(n int, author uint64, replyC chan types.ReplyEvent, auth api.Authenticator, network external.Network, logger external.Logger) *reliableLogImpl {
	logger.Noticef("Init log manager for replica %d", author)
	re := make(map[uint64]*recorder)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		re[id] = newRecorder(author, id, auth, logger)
	}
	return &reliableLogImpl{
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

func (rl *reliableLogImpl) generate(bid *commonProto.BatchId) {
	event := types.RecvEvent{
		EventType: types.LogRecvGenerate,
		Event:     bid,
	}
	rl.recvC <- event
}

func (rl *reliableLogImpl) record(msg *commonProto.SignedMsg) {
	event := types.RecvEvent{
		EventType: types.LogRecvRecord,
		Event:     msg,
	}
	rl.recvC <- event
}

func (rl *reliableLogImpl) ready(tag *commonProto.BinaryTag) {
	event := types.RecvEvent{
		EventType: types.LogRecvReady,
		Event:     tag,
	}
	rl.recvC <- event
}

func (rl *reliableLogImpl) start() {
	go rl.listener()
}

func (rl *reliableLogImpl) stop() {
	select {
	case <-rl.closeC:
	default:
		close(rl.closeC)
	}
}

func (rl *reliableLogImpl) listener() {
	for {
		select {
		case <-rl.closeC:
			rl.logger.Notice("exist log manager listener")
			return
		case ev := <-rl.recvC:
			rl.dispatchRecvEvent(ev)
		}
	}
}

func (rl *reliableLogImpl) dispatchRecvEvent(event types.RecvEvent) {
	switch event.EventType {
	case types.LogRecvGenerate:
		bid, ok := event.Event.(*commonProto.BatchId)
		if !ok {
			return
		}
		msg := rl.generator.generate(bid)
		bin, ok := rl.binary[msg.Sequence]
		if !ok {
			bin = newBinary(rl.n, rl.author, msg.Sequence, rl.replyC, rl.logger)
			rl.binary[msg.Sequence] = bin
		}
		bin.update(msg)
		rl.recorder[msg.Author].logs[msg.Sequence] = msg
	case types.LogRecvRecord:
		signed, ok := event.Event.(*commonProto.SignedMsg)
		if !ok {
			return
		}
		msg := rl.recorder[signed.Author].update(signed)
		if msg == nil {
			rl.logger.Warningf("something wrong with recorder update")
			return
		}
		bin, ok := rl.binary[msg.Sequence]
		if !ok {
			bin = newBinary(rl.n, rl.author, msg.Sequence, rl.replyC, rl.logger)
			rl.binary[msg.Sequence] = bin
		}
		bin.update(msg)

		if tag := bin.getTag(); tag != nil {
			rl.processBinaryTag(tag)
		}
	case types.LogRecvReady:
		// todo we need to request ready event until we get the sequence of execute logs
		bTag := event.Event.(*commonProto.BinaryTag)

		rl.logger.Infof("replica %d is ready on sequence %d, set %v", rl.author, bTag.Sequence, bTag.BinarySet)

		bin, ok := rl.binary[bTag.Sequence]
		if !ok || !bin.finished {
			rl.logger.Warningf("replica %d cannot trigger ready for sequence %d", rl.author, bTag.Sequence)
			return
		}

		bin.ready(bTag)
		rl.processBinaryTag(bTag)
	default:
		return
	}
}

func (rl *reliableLogImpl) processBinaryTag(tag *commonProto.BinaryTag) {
	var logs []*commonProto.OrderedMsg
	var missing []uint64

	for index, value := range tag.BinarySet {
		id := uint64(index+1)
		if value == 1 {
			log, ok := rl.recorder[id].logs[tag.Sequence]
			if !ok {
				missing = append(missing, id)
			}
			logs = append(logs, log)
		}
	}

	if len(missing) > 0 {
		rl.logger.Debugf("replica %d miss the logs from %v", rl.author, missing)
		mm := types.MissingMsg{
			Tag:       tag,
			MissingID: missing,
		}
		event := types.ReplyEvent{
			EventType: types.LogReplyMissingEvent,
			Event:     mm,
		}
		rl.replyC <- event
	} else {
		for _, re := range rl.recorder {
			re.upgrade(tag.Sequence)
		}

		event := types.ReplyEvent{
			EventType: types.LogReplyExecuteEvent,
			Event:     logs,
		}
		rl.replyC <- event
	}
}
