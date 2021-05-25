package phalanx

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/core/types"
)

func NewPhalanx(conf types.Config) *phalanxImpl {
	return newPhalanxImpl(conf)
}

func (phi *phalanxImpl) Start() {
	phi.start()
}

func (phi *phalanxImpl) Stop() {
	phi.stop()
}

func (phi *phalanxImpl) IsNormal() bool {
	return !phi.txpool.IsPoolFull()
}

func (phi *phalanxImpl) PostTxs(txs []*commonProto.Transaction) {
	phi.postTxs(txs)
}

func (phi *phalanxImpl) Execute(payload []byte) {
	phi.execute(payload)
}
