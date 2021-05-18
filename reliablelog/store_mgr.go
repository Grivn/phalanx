package reliablelog

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type orderedIdx struct {
	// author indicates the node generates current log
	author uint64

	// sequence indicates the seqNo
	sequence uint64
}

type orderedValue struct {
	// id indicates which request is current log for
	log *commonProto.OrderedLog

	// ack indicates the nodes received current log, key - verifier, value - existence
	ack map[uint64]bool

	// send is used to track if current replica has approved this ordered-log
	send bool

	// trusted indicates if there are f+1 replicas agreed on current ordered-val
	trusted bool

	// stable indicates if there are n-f replicas agreed on current ordered-val
	stable bool
}

type orderedVerifier struct {
	// n and f are the basic values for byzantine fault tolerance
	n int
	f int

	// author indicates the identifier of current node
	author uint64

	sendC commonTypes.ReliableSendChan

	sender external.Network

	// logMap is used to track the messages about ordered-log and ack
	// when we generate or receive a log, we would like to put it into such a structure
	// and we will use it to track the ack messages about it
	logMap map[orderedIdx]*orderedValue

	// logger is used to print logs
	logger external.Logger
}

func newOrderedVerifier(n int, author uint64, sendC commonTypes.ReliableSendChan, network external.Network, logger external.Logger) *orderedVerifier {
	return &orderedVerifier{
		n:      n,
		f:      (n-1)/3,
		sendC:  sendC,
		author: author,
		sender: network,
		logMap: make(map[orderedIdx]*orderedValue),
		logger: logger,
	}
}

func (ov *orderedVerifier) recordLog(log *commonProto.OrderedLog) {
	val := ov.getOrderedVal(log.Author, log.Sequence, log)

	if !ov.compare() {
		// todo compare not equal
		ov.logger.Error("log compare not equal")
		return
	}

	if !val.send {
		ov.sendAck(val.log)
		val.send = true
	}

	ov.checkValue(val)
}

func (ov *orderedVerifier) recordAck(ack *commonProto.OrderedAck) {
	val := ov.getOrderedVal(ack.OrderedLog.Author, ack.OrderedLog.Sequence, ack.OrderedLog)

	if !ov.compare() {
		// todo compare not equal
		ov.logger.Error("ack compare not equal")
		return
	}
	val.ack[ack.Author] = true

	ov.checkValue(val)
}

func (ov *orderedVerifier) checkValue(val *orderedValue) {
	if len(val.ack) < ov.f+1 {
		return
	}

	if !val.trusted {
		ov.logger.Debugf("Replica %d has received f+1 ack, trusted log for author %d sequence %d", ov.author, val.log.Author, val.log.Sequence)
		ov.sendC.TrustedChan <- val.log
		val.trusted = true
	}

	if !val.send {
		ov.sendAck(val.log)
		val.send = true
	}

	if len(val.ack) < ov.n-ov.f {
		return
	}

	if !val.stable {
		ov.logger.Debugf("Replica %d has received n-f ack, stable log for author %d sequence %d", ov.author, val.log.Author, val.log.Sequence)
		ov.sendC.StableChan <- val.log
		val.stable = true
	}
}

func (ov *orderedVerifier) sendAck(log *commonProto.OrderedLog) {
	ov.logger.Debugf("Replica %d send ack for author %d sequence %d", ov.author, log.Author, log.Sequence)

	ack := &commonProto.OrderedAck{
		Author:     ov.author,
		OrderedLog: log,
	}
	ov.sender.BroadcastAck(ack)
}

func (ov *orderedVerifier) getOrderedVal(author, sequence uint64, log *commonProto.OrderedLog) *orderedValue {
	val, ok := ov.logMap[orderedIdx{author: author, sequence: sequence}]
	if !ok {
		val = &orderedValue{
			log: log,
			ack: make(map[uint64]bool),
		}
		ov.logMap[orderedIdx{author: author, sequence: sequence}] = val
	}
	return val
}

func (ov *orderedVerifier) compare() bool {
	return true
}

func (ov *orderedVerifier) remove(replicaID uint64, seqNo uint64) {
	delete(ov.logMap, orderedIdx{author: replicaID, sequence: seqNo})
}
