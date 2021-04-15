package txpool

import (
	"github.com/Grivn/phalanx/api"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewTxPool(author uint64, batchSize, poolSize int, batched chan *commonProto.Batch, executor external.Executor, network external.Network, logger external.Logger) api.TxPool {
	return newTxPoolImpl(author, batchSize, poolSize, batched, executor, network, logger)
}

func (tp *txPoolImpl) Start() {
	tp.start()
}

func (tp *txPoolImpl) Stop() {
	tp.stop()
}

func (tp *txPoolImpl) Reset() {
	tp.reset()
}

func (tp *txPoolImpl) PostTx(tx *commonProto.Transaction) {
	tp.postTx(tx)
}

func (tp *txPoolImpl) PostBatch(batch *commonProto.Batch) {
	tp.postBatch(batch)
}

func (tp *txPoolImpl) ExecuteBlock(block *commonTypes.Block) {
	tp.executeBlock(block)
}

func (tp *txPoolImpl) IsPoolFull() bool {
	return tp.isPoolFull()
}
