package phalanx

import (
	"github.com/Grivn/phalanx/api"
	"github.com/Grivn/phalanx/binsubset"
	ss "github.com/Grivn/phalanx/binsubset/types"
	commonTypes "github.com/Grivn/phalanx/common/types"
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
	bbyC   chan ss.ReplyEvent
	exeC   chan ex.ReplyEvent
	closeC chan bool

	logger external.Logger
}

func newPhalanxImpl(n int, author uint64, auth api.Authenticator, exec external.Executor, network external.Network, logger external.Logger) *phalanxImpl {
	logger.Noticef("[INIT] replica %d init phalanx consensus protocol", author)

	txpC := make(chan tp.ReplyEvent)
	reqC := make(chan re.ReplyEvent)
	logC := make(chan rl.ReplyEvent)
	bbyC := make(chan ss.ReplyEvent)
	exeC := make(chan ex.ReplyEvent)

	if auth == nil {
		logger.Error("nil authentication")
		return nil
	}

	return &phalanxImpl{
		author: author,

		txpool:      txpool.NewTxPool(author, 100, txpC, exec, network, logger),
		requester:   requester.NewRequester(n, author, reqC, network, logger),
		reliableLog: reliablelog.NewReliableLog(n, author, logC, auth, network, logger),
		byzantine:   binsubset.NewSubset(n, author, bbyC, network, logger),
		executor:    executor.NewExecutor(n, author, exeC, logger),

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
			phi.dispatchCommMsg(comm)
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
		err := proto.Unmarshal(comm.Payload, batch)
		if err != nil {
			phi.logger.Errorf("Unmarshal error: %s", err)
			return
		}
		phi.txpool.PostBatch(batch)
	case commonProto.CommType_ORDER:
		msg := &commonProto.OrderedMsg{}
		err := proto.Unmarshal(comm.Payload, msg)
		if err != nil {
			phi.logger.Errorf("Unmarshal error: %s", err)
			return
		}
		phi.requester.Record(msg)
	case commonProto.CommType_SIGNED:
		signed := &commonProto.SignedMsg{}
		err := proto.Unmarshal(comm.Payload, signed)
		if err != nil {
			phi.logger.Errorf("Unmarshal error: %s", err)
			return
		}
		phi.reliableLog.Record(signed)
	case commonProto.CommType_BBA:
		ntf := &commonProto.BinaryNotification{}
		err := proto.Unmarshal(comm.Payload, ntf)
		if err != nil {
			phi.logger.Errorf("Unmarshal error: %s", err)
			return
		}
		phi.byzantine.Propose(ntf)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchTxPoolEvent(event tp.ReplyEvent) {
	switch event.EventType {
	case tp.ReplyGenerateBatchEvent:
		batch, ok := event.Event.(*commonProto.Batch)
		if !ok {
			phi.logger.Error("parsing error")
			return
		}
		phi.requester.Generate(batch.BatchId)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchRequestEvent(event re.ReplyEvent) {
	switch event.EventType {
	case re.ReqReplyBatchByOrder:
		bid, ok := event.Event.(*commonProto.BatchId)
		if !ok {
			phi.logger.Error("parsing error")
			return
		}
		phi.reliableLog.Generate(bid)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchLogEvent(event rl.ReplyEvent) {
	switch event.EventType {
	case rl.LogReplyQuorumBinaryEvent:
		tag, ok := event.Event.(*commonProto.BinaryTag)
		if !ok {
			phi.logger.Error("parsing error")
			return
		}
		phi.byzantine.Trigger(tag)
	case rl.LogReplyExecuteEvent:
		exec, ok := event.Event.(rl.ExecuteLogs)
		if !ok {
			phi.logger.Error("parsing error")
			return
		}
		phi.executor.ExecuteLogs(exec.Sequence, exec.Logs)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchByzantineEvent(event ss.ReplyEvent) {
	switch event.EventType {
	case ss.BinaryReplyReady:
		tag, ok := event.Event.(*commonProto.BinaryTag)
		if !ok {
			phi.logger.Error("parsing error")
			return
		}
		phi.reliableLog.Ready(tag)
	default:
		return
	}
}

func (phi *phalanxImpl) dispatchExecutorEvent(event ex.ReplyEvent) {
	switch event.EventType {
	case ex.ExecReplyExecuteBlock:
		blk, ok := event.Event.(*commonTypes.Block)
		if !ok {
			phi.logger.Error("parsing error")
			return
		}
		phi.txpool.ExecuteBlock(blk)
	default:
		return
	}
}
