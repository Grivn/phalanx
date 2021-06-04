package phalanx

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/core/types"
	"github.com/Grivn/phalanx/executor"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/remotelog"
	"github.com/Grivn/phalanx/requester"
	"github.com/Grivn/phalanx/txpool"
	"github.com/gogo/protobuf/proto"
)

type phalanxImpl struct {
	n int
	f int
	author uint64

	txpool    internal.TxPool
	requester internal.Requester
	reliable  internal.ReliableLog
	executor  internal.Executor

	commC  commonTypes.CommChan
	txpC   commonTypes.TxPoolSendChan
	reqC   commonTypes.RequesterSendChan
	exeC   commonTypes.ExecutorSendChan
	closeC chan bool

	logger external.Logger
}

func newPhalanxImpl(conf types.Config) *phalanxImpl {
	conf.Logger.Noticef("[INIT] replica %d init phalanx consensus protocol", conf.Author)

	txpC := commonTypes.TxPoolSendChan{
		BatchedChan: make(chan *commonProto.Batch),
	}
	reqC := commonTypes.RequesterSendChan {
		BatchIdChan: make(chan *commonProto.BatchId),
	}
	exeC := commonTypes.ExecutorSendChan{
		BlockChan: make(chan *commonTypes.Block),
	}

	return &phalanxImpl{
		n: conf.N,
		f: (conf.N-1)/3,
		author: conf.Author,

		txpool:    txpool.NewTxPool(conf.Author, conf.BatchSize, conf.PoolSize, txpC, conf.Executor, conf.Network, conf.Logger),
		requester: requester.NewRequester(conf.N, conf.Author, reqC, conf.Network, conf.Logger),
		reliable:  reliablelog.NewReliableLog(conf.N, conf.Author, conf.ReliableC, conf.Network, conf.Logger),
		executor:  executor.NewExecutor(conf.N, conf.Author, exeC, conf.Logger),

		commC:  conf.CommC,
		txpC:   txpC,
		reqC:   reqC,
		exeC:   exeC,
		closeC: make(chan bool),

		logger: conf.Logger,
	}
}

func (phi *phalanxImpl) start() {
	phi.txpool.Start()
	phi.requester.Start()
	phi.reliable.Start()
	phi.executor.Start()

	go phi.listenBatch()
	go phi.listenReq()
	go phi.listenLog()
	go phi.listenAck()

	go phi.listenTxPool()
	go phi.listenRequester()
	go phi.listenExecutor()
}

func (phi *phalanxImpl) stop() {
	phi.txpool.Stop()
	phi.requester.Stop()
	phi.reliable.Stop()
	phi.executor.Stop()

	select {
	case <-phi.closeC:
	default:
		close(phi.closeC)
	}
}

func (phi *phalanxImpl) postTxs(txs []*commonProto.Transaction) {
	phi.logger.Infof("Replica %d received transferred txs from internal", phi.author)
	for _, tx := range txs {
		phi.txpool.PostTx(tx)
	}
}

func (phi *phalanxImpl) listenBatch() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist communicate message listener for phalanx")
			return
		case batch, ok := <-phi.commC.BatchChan:
			if !ok {
				continue
			}
			phi.txpool.PostBatch(batch)
		}
	}
}

func (phi *phalanxImpl) listenReq() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist communicate message listener for phalanx")
			return
		case proposal, ok := <-phi.commC.PropChan:
			if !ok {
				continue
			}
			phi.txpool.PostBatch(proposal.TxBatch)
			phi.requester.Record(req)
		}
	}
}

func (phi *phalanxImpl) listenLog() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist communicate message listener for phalanx")
			return
		case log, ok := <-phi.commC.LogChan:
			if !ok {
				continue
			}
			phi.reliable.RecordLog(log)
		}
	}
}

func (phi *phalanxImpl) listenAck() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist communicate message listener for phalanx")
			return
		case ack, ok := <-phi.commC.AckChan:
			if !ok {
				continue
			}
			phi.reliable.RecordAck(ack)
		}
	}
}

func (phi *phalanxImpl) listenTxPool() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist tx pool listener for phalanx")
			return
		case batch, ok := <-phi.txpC.BatchedChan:
			if !ok {
				continue
			}
			phi.requester.Generate(batch.BatchId)
		}
	}
}

func (phi *phalanxImpl) listenRequester() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist request manager listener for phalanx")
			return
		case bid, ok := <-phi.reqC.BatchIdChan:
			if !ok {
				continue
			}
			phi.reliable.Generate(bid)
		}
	}
}

func (phi *phalanxImpl) execute(payload []byte) {
	exec := &commonProto.ExecuteLogs{}
	err := proto.Unmarshal(payload, exec)
	if err != nil {
		phi.logger.Errorf("Unmarshal error: %s", err)
		return
	}
	phi.executor.ExecuteLogs(exec)
}

func (phi *phalanxImpl) listenExecutor() {
	for {
		select {
		case <-phi.closeC:
			phi.logger.Notice("exist executor listener for phalanx")
			return
		case blk, ok := <-phi.exeC.BlockChan:
			if !ok {
				continue
			}
			phi.txpool.ExecuteBlock(blk)
		}
	}
}

func (phi *phalanxImpl) allCorrectQuorum() int {
	return phi.n-phi.f
}
