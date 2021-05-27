package txpool

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

func NewTxPool(author uint64, batchSize, poolSize int, sendC commonTypes.TxPoolSendChan, executor external.Executor, network external.Network, logger external.Logger) internal.TxPool {
	return newTxPoolImpl(author, batchSize, poolSize, sendC, executor, logger)
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

func (tp *txPoolImpl) PostTx(txs []*commonProto.Transaction) {
	tp.receiveTransactions(txs)
}

func (tp *txPoolImpl) PostBatch(batch *commonProto.TxBatch) {
	tp.receiveTxBatch(batch)
}

func (tp *txPoolImpl) ExecuteBlock(block *commonTypes.Block) {
	tp.tryingBlockExecution(block)
}

func (tp *txPoolImpl) IsPoolFull() bool {
	return tp.isPoolFull()
}
