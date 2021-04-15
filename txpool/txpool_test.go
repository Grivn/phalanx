package txpool

import (
	"sync"
	"testing"
	"time"

	"github.com/Grivn/phalanx/common/mocks"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/stretchr/testify/assert"
)

func TestNewTxPool(t *testing.T) {
	network := mocks.NewFakeNetwork()
	logger := mocks.NewRawLogger()
	batchedC := make(chan *commonProto.Batch)
	txpool := newTxPoolImpl(uint64(1), 100, 50000, batchedC, nil, network, logger)
	assert.Equal(t, 100, txpool.batchSize)
}

func TestTxPoolImpl_Basic(t *testing.T) {
	batchSize := 100
	poolSize := 50000
	network := mocks.NewFakeNetwork()
	logger := mocks.NewRawLogger()

	batchedC1 := make(chan *commonProto.Batch)
	txpool1 := NewTxPool(uint64(1), batchSize, poolSize, batchedC1, nil, network, logger)

	batchedC2 := make(chan *commonProto.Batch)
	txpool2 := NewTxPool(uint64(1), batchSize, poolSize, batchedC2, nil, network, logger)

	txpool1.Start()
	txpool2.Start()

	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	var batch1 *commonProto.Batch

	// replica 1 generate batch
	wg1.Add(2)
	go func() {
		batch1 = <-batchedC1
		wg1.Done()
	}()
	go func() {
		for i:= 0; i< batchSize; i++ {
			tx := mocks.NewTx()
			txpool1.PostTx(tx)
		}
		wg1.Done()
	}()
	wg1.Wait()

	// replica 2 receive the batch from replica 1 and load it
	wg2.Add(1)
	go func() {
		txpool2.PostBatch(batch1)
		wg2.Done()
	}()
	wg2.Wait()

	// replica 2 receive a illegal batch and load it
	tx := mocks.NewTx()
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

	txpool1.Reset()
	txpool2.Reset()

	txpool1.Stop()
	txpool2.Stop()
}
