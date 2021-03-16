package phalanx

import (
	"fmt"
	"github.com/Grivn/phalanx/api"
	"github.com/Grivn/phalanx/binbyzantine"
	bb "github.com/Grivn/phalanx/binbyzantine/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/executor"
	ex "github.com/Grivn/phalanx/executor/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/reliablelog"
	rl "github.com/Grivn/phalanx/reliablelog/types"
	"github.com/Grivn/phalanx/requester"
	re "github.com/Grivn/phalanx/requester/types"
	"github.com/Grivn/phalanx/txpool"
	tp "github.com/Grivn/phalanx/txpool/types"
	"github.com/gogo/protobuf/proto"
)

type phalanxImpl struct {
	author uint64

	txpool      api.TxPool
	requester   api.Requester
	reliableLog api.ReliableLog
	byzantine   api.BinaryByzantine
	executor    api.Executor

	recvC  chan *commonProto.CommMsg
	txpC   chan tp.ReplyEvent
	reqC   chan re.ReplyEvent
	logC   chan rl.ReplyEvent
	bbyC   chan bb.ReplyEvent
	exeC   chan ex.ReplyEvent
	closeC chan bool

	logger external.Logger
}

func newPhalanxImpl(n int, author uint64, auth api.Authenticator, exec external.Executor, network external.Network, logger external.Logger) *phalanxImpl {

	txpC := make(chan tp.ReplyEvent)
	reqC := make(chan re.ReplyEvent)
	logC := make(chan rl.ReplyEvent)
	bbyC := make(chan bb.ReplyEvent)
	exeC := make(chan ex.ReplyEvent)

	txpoolConfig := tp.Config{
		Author:  author,
		Size:    100,
		ReplyC:  txpC,
		Network: network,
		Logger:  logger,
	}

	return &phalanxImpl{
		author: author,

		txpool:      txpool.NewTxPool(txpoolConfig),
		requester:   requester.NewRequester(n, author, reqC, network, logger),
		reliableLog: reliablelog.NewReliableLog(n, author, logC, auth, network, logger),
		byzantine:   binbyzantine.NewByzantine(n, author, bbyC, network, logger),
		executor:    executor.NewExecutor(n, author, exeC, exec, logger),

		recvC:  make(chan *commonProto.CommMsg),
		txpC:   txpC,
		reqC:   reqC,
		logC:   logC,
		bbyC:   bbyC,
		exeC:   exeC,
		closeC: make(chan bool),

		logger: logger,
	}
}

func (phi *phalanxImpl) start() {
	phi.txpool.Start()
	phi.requester.Start()
	phi.reliableLog.Start()
	phi.byzantine.Start()
	phi.executor.Start()

	go phi.listenTxPool()
	go phi.listenRequester()
	go phi.listenReliableLog()
	go phi.listenByzantine()
	go phi.listenExecutor()

	go phi.listenCommMsg()
}

func (phi *phalanxImpl) stop() {
	phi.txpool.Stop()
	phi.requester.Stop()
	phi.reliableLog.Stop()
	phi.byzantine.Stop()
	phi.executor.Stop()

	select {
	case <-phi.closeC:
	default:
		close(phi.closeC)
	}
}

func (phi *phalanxImpl) postTxs(txs []*commonProto.Transaction) {
	phi.logger.Infof("Replica %d received transferred txs from api", phi.author)
	for _, tx := range txs {
		phi.txpool.PostTx(tx)
	}
}

func (phi *phalanxImpl) propose(comm *commonProto.CommMsg) {
	phi.recvC <- comm
}

func (phi *phalanxImpl) listenCommMsg() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist communicate message listener for phalanx")
			return
		case comm := <-phi.recvC:
			fmt.Println(comm)
		}
	}
}

func (phi *phalanxImpl) listenTxPool() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist tx pool listener for phalanx")
			return
		case ev := <-phi.txpC:
			phi.dispatchTxPoolEvent(ev)
		}
	}
}

func (phi *phalanxImpl) listenRequester() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist request manager listener for phalanx")
			return
		case ev := <-phi.reqC:
			phi.dispatchRequestEvent(ev)
		}
	}
}

func (phi *phalanxImpl) listenReliableLog() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist log manager listener for phalanx")
			return
		case ev := <-phi.logC:
			phi.dispatchLogEvent(ev)
		}
	}
}

func (phi *phalanxImpl) listenByzantine() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist log manager listener for phalanx")
			return
		case ev := <-phi.bbyC:
			phi.dispatchByzantineEvent(ev)
		}
	}
}

func (phi *phalanxImpl) listenExecutor() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist executor listener for phalanx")
			return
		case ev := <-phi.exeC:
			phi.dispatchExecutorEvent(ev)
		}
	}
}

func (phi *phalanxImpl) dispatchCommMsg(comm *commonProto.CommMsg) {
	switch comm.Type {
	case commonProto.CommType_BATCH:
		batch := &commonProto.Batch{}
		_ = proto.Unmarshal(comm.Payload, batch)
		phi.txpool.PostBatch(batch)
	case commonProto.CommType_ORDER:
		msg := &commonProto.OrderedMsg{}
		_ = proto.Unmarshal(comm.Payload, msg)
		phi.requester.Record(msg)
	case commonProto.CommType_SIGNED:
		signed := &commonProto.SignedMsg{}
		_ = proto.Unmarshal(comm.Payload, signed)
		phi.reliableLog.Record(signed)
	case commonProto.CommType_BBA:
		ntf := &commonProto.BinaryNotification{}
		_ = proto.Unmarshal(comm.Payload, ntf)
		phi.byzantine.Propose(ntf)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchTxPoolEvent(event tp.ReplyEvent) {
	switch event.EventType {
	case tp.ReplyGenerateBatchEvent:
		batch := event.Event.(*commonProto.Batch)
		phi.requester.Generate(batch.BatchId)
	case tp.ReplyLoadBatchEvent:
		batch := event.Event.(*commonProto.Batch)
		phi.executor.ExecuteBatch(batch)
	case tp.ReplyMissingBatchEvent:
		// todo skip
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchRequestEvent(event re.ReplyEvent) {
	switch event.EventType {
	case re.ReqReplyBatchByOrder:
		bid := event.Event.(*commonProto.BatchId)
		phi.reliableLog.Generate(bid)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchLogEvent(event rl.ReplyEvent) {
	switch event.EventType {
	case rl.LogReplyQuorumBinaryEvent:
		tag := event.Event.(*commonProto.BinaryTag)
		phi.byzantine.Trigger(tag)
	case rl.LogReplyExecuteEvent:
		exec := event.Event.(rl.ExecuteLogs)
		phi.executor.ExecuteLogs(exec.Sequence, exec.Logs)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchByzantineEvent(event bb.ReplyEvent) {
	switch event.EventType {
	case bb.BinaryReplyReady:
		tag := event.Event.(*commonProto.BinaryTag)
		phi.reliableLog.Ready(tag)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchExecutorEvent(event ex.ReplyEvent) {
	switch event.EventType {
	case ex.ExecReplyLoadBatch:
		bid := event.Event.(*commonProto.BatchId)
		phi.txpool.Load(bid)
	default:
		return
	}
}
