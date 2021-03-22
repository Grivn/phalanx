package loggenerator

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/loggenerator/types"
)

type logGeneratorImpl struct {
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

	logger external.Logger
}

func newLogGeneratorImpl(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) *logGeneratorImpl {
	logger.Noticef("Init log manager for replica %d", author)
	re := make(map[uint64]*recorder)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		re[id] = newRecorder(author, id, logger)
	}
	return &logGeneratorImpl{
		n:         n,
		author:    author,
		recvC:     make(chan types.RecvEvent),
		replyC:    replyC,
		closeC:    make(chan bool),
		generator: newGenerator(author, replyC, re[author], network, logger),
		recorder:  re,
		binary:    make(map[uint64]*binary),
		logger:    logger,
	}
}

func (lg *logGeneratorImpl) generate(bid *commonProto.BatchId) {
	event := types.RecvEvent{
		EventType: types.LogRecvGenerate,
		Event:     bid,
	}
	lg.recvC <- event
}

func (lg *logGeneratorImpl) record(msg *commonProto.OrderedMsg) {
	event := types.RecvEvent{
		EventType: types.LogRecvRecord,
		Event:     msg,
	}
	lg.recvC <- event
}

func (lg *logGeneratorImpl) ready(tag *commonProto.BinaryTag) {
	event := types.RecvEvent{
		EventType: types.LogRecvReady,
		Event:     tag,
	}
	lg.recvC <- event
}

func (lg *logGeneratorImpl) start() {
	go lg.listener()
}

func (lg *logGeneratorImpl) stop() {
	select {
	case <-lg.closeC:
	default:
		close(lg.closeC)
	}
}

func (lg *logGeneratorImpl) listener() {
	for {
		select {
		case <-lg.closeC:
			lg.logger.Notice("exist log manager listener")
			return
		case ev := <-lg.recvC:
			lg.dispatchRecvEvent(ev)
		}
	}
}

func (lg *logGeneratorImpl) dispatchRecvEvent(event types.RecvEvent) {
	switch event.EventType {
	case types.LogRecvGenerate:
		bid, ok := event.Event.(*commonProto.BatchId)
		if !ok {
			return
		}
		msg := lg.generator.generate(bid)
		bin, ok := lg.binary[msg.Sequence]
		if !ok {
			bin = newBinary(lg.n, lg.author, msg.Sequence, lg.replyC, lg.logger)
			lg.binary[msg.Sequence] = bin
		}
		bin.update(msg)
		bin.onActive()
		lg.recorder[msg.Author].logs[msg.Sequence] = msg

		if tag := bin.getTag(); tag != nil {
			lg.processBinaryTag(tag)
		}
	case types.LogRecvRecord:
		msg, ok := event.Event.(*commonProto.OrderedMsg)
		if !ok {
			return
		}
		if !lg.recorder[msg.Author].update(msg) {
			lg.logger.Warningf("something wrong with recorder update")
			return
		}
		bin, ok := lg.binary[msg.Sequence]
		if !ok {
			bin = newBinary(lg.n, lg.author, msg.Sequence, lg.replyC, lg.logger)
			lg.binary[msg.Sequence] = bin
		}
		bin.update(msg)

		if !bin.isActive() {
			lg.logger.Warningf("replica %d binary processor for sequence %d has not started", lg.author, msg.Sequence)
			return
		}

		if tag := bin.getTag(); tag != nil {
			lg.processBinaryTag(tag)
		}
	case types.LogRecvReady:
		// todo we need to request ready event until we get the sequence of execute logs
		bTag := event.Event.(*commonProto.BinaryTag)

		lg.logger.Infof("replica %d is ready on sequence %d, set %v", lg.author, bTag.Sequence, bTag.BinarySet)

		bin, ok := lg.binary[bTag.Sequence]
		if !ok {
			bin = newBinary(lg.n, lg.author, bTag.Sequence, lg.replyC, lg.logger)
			lg.binary[bTag.Sequence] = bin
		}
		bin.ready(bTag)

		if !bin.isActive() {
			lg.logger.Warningf("replica %d binary processor for sequence %d has not started", lg.author, bTag.Sequence)
			return
		}

		lg.processBinaryTag(bTag)
	default:
		return
	}
}

func (lg *logGeneratorImpl) processBinaryTag(tag *commonProto.BinaryTag) {
	var logs []*commonProto.OrderedMsg
	var missing []uint64

	for index, value := range tag.BinarySet {
		id := uint64(index+1)
		if value == 1 {
			log, ok := lg.recorder[id].logs[tag.Sequence]
			if !ok {
				missing = append(missing, id)
			}
			logs = append(logs, log)
		}
	}

	if len(missing) > 0 {
		lg.logger.Debugf("replica %d miss the logs from %v", lg.author, missing)
		mm := types.MissingMsg{
			Tag:       tag,
			MissingID: missing,
		}
		event := types.ReplyEvent{
			EventType: types.LogReplyMissingEvent,
			Event:     mm,
		}
		lg.replyC <- event
	} else {
		for _, re := range lg.recorder {
			re.upgrade(tag.Sequence)
		}

		exec := types.ExecuteLogs{
			Sequence: tag.Sequence,
			Logs:     logs,
		}

		event := types.ReplyEvent{
			EventType: types.LogReplyExecuteEvent,
			Event:     exec,
		}
		lg.replyC <- event
	}
}
