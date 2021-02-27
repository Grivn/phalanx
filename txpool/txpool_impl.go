package txpool

import (
	"time"

	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/txpool/types"
)

type txPoolImpl struct {
	author uint64

	size int // batch size

	pending *recorder

	batches map[*commonProto.BatchId]*commonProto.Batch

	recvC chan types.RecvEvent

	replyC chan types.ReplyEvent

	close chan bool

	sender *sender

	logger external.Logger
}

func newTxPoolImpl(config types.Config) *txPoolImpl {
	return &txPoolImpl{
		author:  config.Author,
		size:    config.Size,
		pending: newRecorder(),
		batches: make(map[*commonProto.BatchId]*commonProto.Batch),
		recvC:   make(chan types.RecvEvent),
		replyC:  config.ReplyC,
		close:   make(chan bool),
		sender:  newSender(config.Author, config.Network),
		logger:  config.Logger,
	}
}

func (tp *txPoolImpl) start() {
	go tp.listener()
}

func (tp *txPoolImpl) stop() {
	close(tp.close)
}

func (tp *txPoolImpl) reset() {
	tp.pending.reset()
	tp.batches = make(map[*commonProto.BatchId]*commonProto.Batch)
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

func (tp *txPoolImpl) load(bid *commonProto.BatchId) {
	event := types.RecvEvent{
		EventType: types.RecvLoadBatchEvent,
		Event:     bid,
	}
	tp.recvC <- event
}

func (tp *txPoolImpl) listener() {
	for {
		select {
		case event := <-tp.recvC:
			tp.processEvents(event)
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
			return
		}
		tp.processRecvTxEvent(tx)
	case types.RecvRecordBatchEvent:
		batch, ok := event.Event.(*commonProto.Batch)
		if !ok {
			return
		}
		tp.processRecvBatchEvent(batch)
	case types.RecvLoadBatchEvent:
		bid, ok := event.Event.(*commonProto.BatchId)
		if !ok {
			return
		}
		tp.processLoadBatchEvent(bid)
	default:
		return
	}
}

func (tp *txPoolImpl) processRecvTxEvent(tx *commonProto.Transaction) {
	tp.pending.update(tx)

	if tp.pending.len() == tp.size {
		batch := tp.generateBatch()
		if batch == nil {
			tp.logger.Warningf("Replica %d generated a nil batch", tp.author)
			return
		}
		tp.replyBatch(batch)
	}
	return
}

func (tp *txPoolImpl) processRecvBatchEvent(batch *commonProto.Batch) {
	tp.logger.Infof("Replica %d received batch %s from replica %d", tp.author, batch.BatchId.BatchHash, batch.BatchId.Author)
	if !tp.verifyBatch(batch) {
		return
	}
	tp.batches[batch.BatchId] = batch
}

func (tp *txPoolImpl) processLoadBatchEvent(bid *commonProto.BatchId) {
	if bid == nil {
		tp.logger.Warningf("Replica %d received a nil request", tp.author)
		return
	}
	batch, ok := tp.batches[bid]
	if !ok {
		tp.logger.Warningf("Replica %d cannot find batch %s", tp.author, bid.BatchHash)
		tp.replyMissingEvent(bid)
		return
	}
	tp.replyBatch(batch)
}

//===================================
//       essential tools
//===================================

func (tp *txPoolImpl) generateBatch() *commonProto.Batch {
	batch := &commonProto.Batch{
		HashList:  tp.pending.hashes(),
		TxList:    tp.pending.txs(),
		Timestamp: time.Now().UnixNano(),
	}

	hash := commonTypes.CalculateMD5Hash(batch.HashList, batch.Timestamp)
	batch.BatchId = &commonProto.BatchId{
		Author:    tp.author,
		BatchHash: commonTypes.BytesToString(hash),
	}

	tp.pending.reset()
	tp.batches[batch.BatchId] = batch
	tp.logger.Noticef("Replica %d generate a batch %s", tp.author, batch.BatchId.BatchHash)
	tp.sender.broadcast(batch)
	return batch
}

func (tp *txPoolImpl) verifyBatch(batch *commonProto.Batch) bool {
	var hashList []string
	for _, tx := range batch.TxList {
		hashList = append(hashList, commonTypes.GetHash(tx))
	}
	batchHash := commonTypes.BytesToString(commonTypes.CalculateMD5Hash(hashList, batch.Timestamp))

	if batchHash != batch.BatchId.BatchHash {
		tp.logger.Warningf("Replica %d received a batch with miss-matched hash from replica %d", tp.author, batch.BatchId.Author)
		return false
	}
	return true
}

//===================================
//       reply messages
//===================================
func (tp *txPoolImpl) replyBatch(batch *commonProto.Batch) {
	reply := types.ReplyEvent{
		EventType: types.ReplyBatchEvent,
		Event:     batch,
	}
	tp.replyC <- reply
}

func (tp *txPoolImpl) replyMissingEvent(id *commonProto.BatchId) {
	reply := types.ReplyEvent{
		EventType: types.ReplyMissingBatchEvent,
		Event:     id,
	}
	tp.replyC <- reply
}
