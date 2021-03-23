package txpool

import (
	"container/list"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/api"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/timer"
	timerTypes "github.com/Grivn/phalanx/timer/types"
	"github.com/Grivn/phalanx/txpool/types"
)

type txPoolImpl struct {
	author uint64

	isFull uint32

	size int // batch size

	blockList *list.List

	pendingTxs *recorder

	batches map[commonTypes.LogID]batchEntry

	recvC chan interface{}

	replyC chan types.ReplyEvent

	close chan bool

	sender *sender

	pendingBlock *pendingBlock

	executor external.Executor

	timer api.Timer

	logger external.Logger
}

type batchEntry struct {
	batch *commonProto.Batch
	local bool
}

type pendingBlock struct {
	e         *list.Element
	blk       *commonTypes.Block
	endIndex  int
	txList    []*commonProto.Transaction
	localList []bool
}

func newTxPoolImpl(author uint64, size int, replyC chan types.ReplyEvent, executor external.Executor, network external.Network, logger external.Logger) *txPoolImpl {
	recvC := make(chan interface{})

	fmt.Println(size)

	return &txPoolImpl{
		author:     author,
		size:       1,
		blockList:  list.New(),
		pendingTxs: newRecorder(),
		batches:    make(map[commonTypes.LogID]batchEntry),
		recvC:      recvC,
		replyC:     replyC,
		close:      make(chan bool),
		executor:   executor,
		sender:     newSender(author, network),
		timer:      timer.NewTimer(recvC, logger),
		logger:     logger,
	}
}

func (tp *txPoolImpl) start() {
	go tp.listener()
}

func (tp *txPoolImpl) stop() {
	close(tp.close)
}

func (tp *txPoolImpl) isPoolFull() bool {
	return atomic.LoadUint32(&tp.isFull) == 1
}

func (tp *txPoolImpl) reset() {
	tp.pendingTxs.reset()
	tp.batches = make(map[commonTypes.LogID]batchEntry)
}

func (tp *txPoolImpl) postBatch(batch *commonProto.Batch) {
	event := types.RecvEvent{
		EventType: types.RecvRecordBatchEvent,
		Event:     batch,
	}
	tp.recvC <- event
}

func (tp *txPoolImpl) postTx(tx *commonProto.Transaction) {
	event := types.RecvEvent{
		EventType: types.RecvRecordTxEvent,
		Event:     tx,
	}
	tp.recvC <- event
}

func (tp *txPoolImpl) executeBlock(block *commonTypes.Block) {
	event := types.RecvEvent{
		EventType: types.RecvExecuteBlock,
		Event:     block,
	}
	tp.recvC <- event
}

func (tp *txPoolImpl) listener() {
	for {
		select {
		case event := <-tp.recvC:
			switch ev := event.(type) {
			case types.RecvEvent:
				tp.processEvents(ev)
			case bool:
				tp.tryToGenerate()
				if len(tp.pendingTxs.txList) > 0 {
					tp.timer.StartTimer(timerTypes.TxPoolTimer, true)
				}
			}
		case <-tp.close:
			tp.logger.Notice("exist tx pool listener")
			return
		}
	}
}

//===================================
//       event processor
//===================================

func (tp *txPoolImpl) processEvents(event types.RecvEvent) {
	switch event.EventType {
	case types.RecvRecordTxEvent:
		tx, ok := event.Event.(*commonProto.Transaction)
		if !ok {
			tp.logger.Error("parsing error")
			return
		}
		tp.processRecvTxEvent(tx)
		tp.timer.StartTimer(timerTypes.TxPoolTimer, true)

		if len(tp.batches)*tp.size+len(tp.pendingTxs.txList) > 50000 {
			atomic.StoreUint32(&tp.isFull, 1)
		}
	case types.RecvRecordBatchEvent:
		batch, ok := event.Event.(*commonProto.Batch)
		if !ok {
			tp.logger.Error("parsing error")
			return
		}
		tp.processRecvBatchEvent(batch)

		if len(tp.batches)*tp.size+len(tp.pendingTxs.txList) > 50000 {
			atomic.StoreUint32(&tp.isFull, 1)
		}
	case types.RecvExecuteBlock:
		blk, ok := event.Event.(*commonTypes.Block)
		if !ok {
			tp.logger.Error("parsing error")
			return
		}
		tp.processExecuteBlock(blk)

		if len(tp.batches)*tp.size+len(tp.pendingTxs.txList) < 50000 {
			atomic.StoreUint32(&tp.isFull, 0)
		}
	default:
		return
	}
}

func (tp *txPoolImpl) processRecvTxEvent(tx *commonProto.Transaction) {
	tp.pendingTxs.update(tx)

	if tp.pendingTxs.len() == tp.size {
		tp.tryToGenerate()
	}
	return
}

func (tp *txPoolImpl) tryToGenerate() {
	batch := tp.generateBatch()
	if batch == nil {
		tp.logger.Warningf("Replica %d generated a nil batch", tp.author)
		return
	}
	tp.replyGenerateBatch(batch)
}

func (tp *txPoolImpl) processRecvBatchEvent(batch *commonProto.Batch) {
	tp.logger.Infof("Replica %d received batch %s from replica %d", tp.author, batch.BatchId.BatchHash, batch.BatchId.Author)
	if !tp.verifyBatch(batch) {
		return
	}

	id := commonTypes.LogID{
		Author: batch.BatchId.Author,
		Hash:   batch.BatchId.BatchHash,
	}
	entry := batchEntry{
		batch: batch,
		local: false,
	}
	tp.batches[id] = entry
}

func (tp *txPoolImpl) processExecuteBlock(blk *commonTypes.Block) {
	tp.blockList.PushBack(blk)
	tp.tryExecuteBlock()
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

	id := commonTypes.LogID{
		Author: tp.author,
		Hash:   hash,
	}
	entry := batchEntry{
		batch: batch,
		local: true,
	}
	tp.batches[id] = entry

	tp.logger.Noticef("replica %d generate a batch %s", tp.author, batch.BatchId.BatchHash)
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
		return
	}

	if tp.pendingBlock == nil {
		e := tp.blockList.Front()
		blk, ok := e.Value.(*commonTypes.Block)
		if !ok {
			tp.logger.Error("parsing block type error")
			return
		}

		tp.pendingBlock = &pendingBlock{
			e:         e,
			blk:       blk,
			endIndex:  0,
			txList:    nil,
			localList: nil,
		}
	}

	blk := tp.pendingBlock.blk
	for index, log := range blk.Logs {
		if index < tp.pendingBlock.endIndex {
			continue
		}

		entry, ok := tp.batches[log.ID]
		if !ok {
			tp.logger.Debugf("replica %d hasn't received batch %v for block %d", tp.author, log.ID, blk.Sequence)
			tp.pendingBlock.endIndex = index
			return
		}

		for _, tx := range entry.batch.TxList {
			tp.pendingBlock.txList = append(tp.pendingBlock.txList, tx)
			tp.pendingBlock.localList = append(tp.pendingBlock.localList, entry.local)
		}
	}

	tp.logger.Noticef("======== replica %d call execute, seqNo=%d, timestamp=%d", tp.author, blk.Sequence, blk.Timestamp)
	tp.executor.Execute(tp.pendingBlock.txList, tp.pendingBlock.localList, blk.Sequence, blk.Timestamp)

	// remove the stored batches
	for _, log := range blk.Logs {
		delete(tp.batches, log.ID)
	}

	// remove the executed block
	tp.blockList.Remove(tp.pendingBlock.e)
	tp.pendingBlock = nil
}

//===================================
//       reply messages
//===================================
func (tp *txPoolImpl) replyGenerateBatch(batch *commonProto.Batch) {
	reply := types.ReplyEvent{
		EventType: types.ReplyGenerateBatchEvent,
		Event:     batch,
	}
	tp.replyC <- reply
}
