package txpool

import (
	"container/list"
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/api"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/timer"
	timerTypes "github.com/Grivn/phalanx/timer/types"
)

// txPoolImpl is the implementation of txPool
type txPoolImpl struct {
	// author indicates the owner of txPool
	author uint64

	// isFull indicates if the txPool full or not
	isFull uint32

	// txPool would generate a batch while the number of txs has reached batchSize
	batchSize int

	// poolSize indicates the maximum size of txPool
	poolSize int

	// pendingTxs is used to record the transactions send from api
	pendingTxs *recorder

	// batchStore is used to record the batches generated by self-node and send from others
	batchStore map[commonTypes.LogID]batchEntry

	// recvC is the channel group which is used to receive events from other modules
	recvC recvChan

	// sendC is the channel group which is used to send back information to other modules
	sendC sendChan

	// timeoutC is used to send the signals for timeout processor (timer)
	timeoutC chan interface{}

	// closeC is used to close the go-routine of txPool
	closeC chan bool

	// sender is used to send consensus message to network
	sender *sender

	// blockList is used to track the blocks waiting for execution
	blockList *list.List

	// pendingBlock is the front block in blockList which will be executed at first
	pendingBlock *pendingBlock

	// executor is used to execute blocks
	executor external.Executor

	// timer is used to process timeout events
	timer api.Timer

	// logger is used to print logs
	logger external.Logger
}

// recvChan is the channel group which is used to receive events from other modules
type recvChan struct {
	// transactionChan is used to receive transactions send from api
	transactionChan chan *commonProto.Transaction

	// batchedChan is used to receive batchStore send from other replicas
	batchedChan chan *commonProto.Batch

	// executeChan is used to receive execute event send from executor module
	executeChan chan *commonTypes.Block
}

// sendChan is the channel group which is used to send back information to other modules
type sendChan struct {
	// batchedChan is used to send back the batch generated by txPool
	batchedChan chan *commonProto.Batch
}

type batchEntry struct {
	batch *commonProto.Batch
	local bool
}

// pendingBlock is the front block in blockList which will be executed at first
type pendingBlock struct {
	// e indicates the element pointer of current block in block list
	e *list.Element

	// blk contains the block information of current block
	blk *commonTypes.Block

	// endIndex is used to note the first log in current block that the batch of log is missed
	endIndex int

	// txList is the txs in current block
	txList []*commonProto.Transaction

	// localList indicates the txs from local or remote
	localList []bool
}

func newTxPoolImpl(author uint64, batchSize, poolSize int, batchedC chan *commonProto.Batch, executor external.Executor, network external.Network, logger external.Logger) *txPoolImpl {
	timeoutC := make(chan interface{})

	recvC := recvChan{
		transactionChan: make(chan *commonProto.Transaction),
		batchedChan:     make(chan *commonProto.Batch),
		executeChan:     make(chan *commonTypes.Block),
	}

	sendC := sendChan{
		batchedChan: batchedC,
	}

	return &txPoolImpl{
		author:     author,
		batchSize:  batchSize,
		poolSize:   poolSize,
		blockList:  list.New(),
		pendingTxs: newRecorder(),
		batchStore: make(map[commonTypes.LogID]batchEntry),

		recvC:    recvC,
		sendC:    sendC,
		timeoutC: timeoutC,
		closeC:   make(chan bool),

		executor:   executor,
		sender:     newSender(author, network),
		timer:      timer.NewTimer(timeoutC, logger),
		logger:     logger,
	}
}

func (tp *txPoolImpl) start() {
	go tp.listener()
}

func (tp *txPoolImpl) stop() {
	close(tp.closeC)
}

func (tp *txPoolImpl) isPoolFull() bool {
	return atomic.LoadUint32(&tp.isFull) == 1
}

func (tp *txPoolImpl) reset() {
	tp.pendingTxs.reset()
	tp.batchStore = make(map[commonTypes.LogID]batchEntry)
}

func (tp *txPoolImpl) postBatch(batch *commonProto.Batch) {
	tp.recvC.batchedChan <- batch
}

func (tp *txPoolImpl) postTx(tx *commonProto.Transaction) {
	tp.recvC.transactionChan <- tx
}

func (tp *txPoolImpl) executeBlock(blk *commonTypes.Block) {
	tp.recvC.executeChan <- blk
}

func (tp *txPoolImpl) listener() {
	for {
		select {
		case tx := <-tp.recvC.transactionChan:
			tp.processRecvTxEvent(tx)
		case batch := <-tp.recvC.batchedChan:
			tp.processRecvBatchEvent(batch)
		case blk := <- tp.recvC.executeChan:
			tp.processExecuteBlock(blk)

		case <-tp.timeoutC:
			tp.processTimeoutEvent()
		default:
			continue
		}
	}
}

//===================================
//       event processor
//===================================

func (tp *txPoolImpl) processRecvTxEvent(tx *commonProto.Transaction) {
	tp.pendingTxs.update(tx)

	// while the length of pending txs has reached batch size, txPool will generate a batch immediately
	if tp.pendingTxs.len() == tp.batchSize {
		// txPool is trying to generate a batch, stop the txPool timer
		tp.timer.StopTimer(timerTypes.TxPoolTimer)
		tp.tryToGenerate()
	}

	if len(tp.pendingTxs.txList) > 0 {
		// there are some pending txs, txPool starts a timer to generate a batch after the interval of timer
		tp.timer.StartTimer(timerTypes.TxPoolTimer, true)
	}
	tp.checkSpace()
}

func (tp *txPoolImpl) tryToGenerate() {
	batch := tp.generateBatch()
	if batch == nil {
		tp.logger.Warningf("Replica %d generated a nil batch", tp.author)
		return
	}
	go tp.sendGeneratedBatch(batch)
}

func (tp *txPoolImpl) processRecvBatchEvent(batch *commonProto.Batch) {
	tp.logger.Infof("Replica %d received batch %s from replica %d", tp.author, batch.BatchId.BatchHash, batch.BatchId.Author)
	if !tp.verifyBatch(batch) {
		tp.logger.Warningf("Replica %d received an illegal batch", tp.author)
		return
	}

	tp.batchStore[commonTypes.LogID{Author: batch.BatchId.Author, Hash: batch.BatchId.BatchHash}] = batchEntry{
		batch: batch,
		local: false,
	}
	tp.checkSpace()
}

func (tp *txPoolImpl) processExecuteBlock(blk *commonTypes.Block) {
	tp.blockList.PushBack(blk)
	tp.tryExecuteBlock()
	tp.checkSpace()
}

//===================================
//       essential tools
//===================================

func (tp *txPoolImpl) generateBatch() *commonProto.Batch {
	batch := &commonProto.Batch{
		HashList:  tp.pendingTxs.hashes(),
		TxList:    tp.pendingTxs.txs(),
		Timestamp: time.Now().UnixNano(),
	}

	hash := commonTypes.CalculateListHash(batch.HashList, batch.Timestamp)
	batch.BatchId = &commonProto.BatchId{
		Author:    tp.author,
		BatchHash: hash,
	}

	tp.pendingTxs.reset()
	tp.batchStore[commonTypes.LogID{Author: tp.author, Hash: hash}] = batchEntry{
		batch: batch,
		local: true,
	}

	tp.logger.Noticef("replica %d generate a batch %s, len %d", tp.author, batch.BatchId.BatchHash, len(batch.TxList))
	tp.sender.broadcast(batch)
	return batch
}

func (tp *txPoolImpl) verifyBatch(batch *commonProto.Batch) bool {
	var hashList []string
	for _, tx := range batch.TxList {
		hashList = append(hashList, commonTypes.GetHash(tx))
	}
	batchHash := commonTypes.CalculateListHash(hashList, batch.Timestamp)

	if batchHash != batch.BatchId.BatchHash {
		tp.logger.Warningf("Replica %d received a batch with mis-matched hash from replica %d", tp.author, batch.BatchId.Author)
		return false
	}
	return true
}

func (tp *txPoolImpl) tryExecuteBlock() {
	if tp.blockList.Len() == 0 {
		// there isn't any block waiting for execution, just return
		return
	}

	// there are some blocks waiting for execution, but we haven't get the information of the first one to execute
	// here, we would like to find the information of the first block to execute
	if tp.pendingBlock == nil {
		e := tp.blockList.Front()
		blk, ok := e.Value.(*commonTypes.Block)
		if !ok {
			tp.logger.Error("parsing block type error")
			return
		}

		tp.pendingBlock = &pendingBlock{
			e:   e,
			blk: blk,
		}
	}

	blk := tp.pendingBlock.blk
	for i:= tp.pendingBlock.endIndex; i< len(blk.Logs); i++ {
		// get the first log in which the batch information hasn't been read
		log := blk.Logs[tp.pendingBlock.endIndex]

		entry, ok := tp.batchStore[log.ID]
		if !ok {
			// we cannot find the batch information just now
			tp.logger.Debugf("replica %d hasn't received batch %v for block %d", tp.author, log.ID, blk.Sequence)
			return
		}

		for _, tx := range entry.batch.TxList {
			tp.pendingBlock.txList = append(tp.pendingBlock.txList, tx)
			tp.pendingBlock.localList = append(tp.pendingBlock.localList, entry.local)
		}
		tp.pendingBlock.endIndex++
	}

	tp.logger.Noticef("======== replica %d call execute, seqNo=%d, timestamp=%d", tp.author, blk.Sequence, blk.Timestamp)
	tp.executor.Execute(tp.pendingBlock.txList, tp.pendingBlock.localList, blk.Sequence, blk.Timestamp)

	// remove the stored batchStore
	for _, log := range blk.Logs {
		delete(tp.batchStore, log.ID)
	}

	// remove the executed block
	tp.blockList.Remove(tp.pendingBlock.e)
	tp.pendingBlock = nil
}

func (tp *txPoolImpl) processTimeoutEvent() {
	if len(tp.pendingTxs.txList) > 0 {
		tp.tryToGenerate()
	}
}

//===================================
//       send back messages
//===================================
func (tp *txPoolImpl) sendGeneratedBatch(batch *commonProto.Batch) {
	tp.sendC.batchedChan <- batch
}

//===================================
//       check txPool full
//===================================
func (tp *txPoolImpl) checkSpace() {
	if len(tp.batchStore)*tp.batchSize+len(tp.pendingTxs.txList) > tp.poolSize {
		atomic.StoreUint32(&tp.isFull, 1)
	} else {
		atomic.StoreUint32(&tp.isFull, 0)
	}
}