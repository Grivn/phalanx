package txpool

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

func NewTxPool(author uint64, batchSize, poolSize int, sendC commonTypes.TxPoolSendChan, executor external.Executor, network external.Network, logger external.Logger) internal.TxPool {
	return newTxPoolImpl(author, batchSize, poolSize, sendC, executor, network, logger)
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

func (tp *txPoolImpl) PostBatch(batch *commonProto.TxBatch) {
	tp.postBatch(batch)
}

func (tp *txPoolImpl) ExecuteBlock(block *commonTypes.Block) {
	tp.executeBlock(block)
}

func (tp *txPoolImpl) IsPoolFull() bool {
	return tp.isPoolFull()
}
