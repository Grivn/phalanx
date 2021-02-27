package txpool

import (
	"sync"
	"testing"
	"time"

	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/common/utils"
	"github.com/Grivn/phalanx/txpool/types"

	"github.com/stretchr/testify/assert"
)

func generateConfig(author uint64, replyC chan types.ReplyEvent) types.Config {
	return types.Config{
		Author:  author,
		Size:    500,
		ReplyC:  replyC,
		Network: utils.NewFakeNetwork(),
		Logger:  utils.NewRawLogger(),
	}
}

func TestNewTxPool(t *testing.T) {
	replyC := make(chan types.ReplyEvent)
	txpool := newTxPoolImpl(generateConfig(uint64(1), replyC))
	assert.Equal(t, 500, txpool.size)
}

func TestTxPoolImpl_Basic(t *testing.T) {
	replyC1 := make(chan types.ReplyEvent)
	config1 := generateConfig(uint64(1), replyC1)
	txpool1 := NewTxPool(config1)

	replyC2 := make(chan types.ReplyEvent)
	config2 := generateConfig(uint64(2), replyC2)
	txpool2 := NewTxPool(config2)

	txpool1.Start()
	txpool2.Start()

	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	var reply1 types.ReplyEvent
	var reply2 types.ReplyEvent

	// replica 1 generate batch
	wg1.Add(2)
	go func() {
		reply1 = <-replyC1
		wg1.Done()
	}()
	go func() {
		for i:= 0; i< config1.Size; i++ {
			tx := utils.NewTx()
			txpool1.PostTx(tx)
		}
		wg1.Done()
	}()
	wg1.Wait()
	assert.Equal(t, types.ReplyBatchEvent, reply1.EventType)
	batch, ok := reply1.Event.(*commonProto.Batch)
	assert.True(t, ok)

	// replica 2 load a missing batch
	wg2.Add(2)
	go func() {
		txpool2.Load(batch.BatchId)
		wg2.Done()
	}()
	go func() {
		reply2 = <-replyC2
		wg2.Done()
	}()
	wg2.Wait()
	assert.Equal(t, types.ReplyMissingBatchEvent, reply2.EventType)

	// replica 2 receive the batch from replica 1 and load it
	wg2.Add(1)
	go func() {
		txpool2.PostBatch(batch)
		wg2.Done()
	}()
	wg2.Wait()
	wg2.Add(2)
	go func() {
		reply2 = <-replyC2
		wg2.Done()
	}()
	go func() {
		txpool2.Load(batch.BatchId)
		wg2.Done()
	}()
	wg2.Wait()
	assert.Equal(t, types.ReplyBatchEvent, reply2.EventType)
	batch2, ok := reply2.Event.(*commonProto.Batch)
	assert.True(t, ok)
	assert.Equal(t, batch.BatchId.BatchHash, batch2.BatchId.BatchHash)

	// replica 2 receive a illegal batch and load it
	tx := utils.NewTx()
	id := &commonProto.BatchId{
		Author:    uint64(1),
		BatchHash: "illegal",
	}
	illegalBatch := &commonProto.Batch{
		BatchId:   id,
		TxList:    []*commonProto.Transaction{tx},
		HashList:  []string{commonTypes.GetHash(tx)},
		Timestamp: time.Now().UnixNano(),
	}
	wg2.Add(1)
	go func() {
		txpool2.PostBatch(illegalBatch)
		wg2.Done()
	}()
	wg2.Wait()
	wg2.Add(2)
	go func() {
		reply2 = <-replyC2
		wg2.Done()
	}()
	go func() {
		txpool2.Load(illegalBatch.BatchId)
		wg2.Done()
	}()
	wg2.Wait()
	assert.Equal(t, types.ReplyMissingBatchEvent, reply2.EventType)

	txpool1.Reset()
	txpool2.Reset()

	txpool1.Stop()
	txpool2.Stop()
}
