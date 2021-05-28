package reliablelog

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type reliableLogImpl struct {
	// author is the current node's identifier
	author uint64

	verifier *orderedVerifier

	// recvC is the channel group which is used to receive events from other modules
	recvC commonTypes.ReliableRecvChan

	// sendC is the channel group which is used to send back information to other modules
	sendC commonTypes.ReliableSendChan

	// closeC is used to close the go-routine of reliableLog
	closeC chan bool

	generator *orderMgr

	logger external.Logger
}

func newReliableLogImpl(n int, author uint64, sendC commonTypes.ReliableSendChan, network external.Network, logger external.Logger) *reliableLogImpl {
	logger.Noticef("replica %d init log manager", author)

	recvC := commonTypes.ReliableRecvChan{
		BatchChan:   make(chan *commonProto.TxBatch),
		PreOrdering: make(chan *commonProto.PreOrdering),
		Vote:        make(chan *commonProto.Vote),
		Ordering:    make(chan *commonProto.Ordering),
	}

	return &reliableLogImpl{
		author:    author,
		verifier:  newOrderedVerifier(n, author, sendC, network, logger),
		recvC:     recvC,
		sendC:     sendC,
		closeC:    make(chan bool),
		generator: newOrderMgr(author, logger),
		logger:    logger,
	}
}

func (rl *reliableLogImpl) generate(batch *commonProto.TxBatch) {

}

func (rl *reliableLogImpl) recordLog(log *commonProto.OrderedLog) {
	rl.recvC.LogChan <- log
}

func (rl *reliableLogImpl) recordAck(ack *commonProto.OrderedAck) {
	rl.recvC.AckChan <- ack
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
		case bid := <-rl.recvC.BatchIdChan:
			rl.verifier.recordLog(rl.generator.generate(bid))
		case log := <-rl.recvC.LogChan:
			rl.verifier.recordLog(log)
		case ack := <-rl.recvC.AckChan:
			rl.verifier.recordAck(ack)
		default:
			continue
		}
	}
}
