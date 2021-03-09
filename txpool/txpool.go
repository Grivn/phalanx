package txpool

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/txpool/types"
)

func NewTxPool(config types.Config) api.TxPool {
	return newTxPoolImpl(config)
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

func (tp *txPoolImpl) Load(bid *commonProto.BatchId) {
	tp.load(bid)
}


