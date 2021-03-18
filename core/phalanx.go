package phalanx

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewPhalanx(n int, author uint64, auth api.Authenticator, exec external.Executor, network external.Network, logger external.Logger) *phalanxImpl {
	return newPhalanxImpl(n, author, auth, exec, network, logger)
}

func (phi *phalanxImpl) Start() {
	phi.start()
}

func (phi *phalanxImpl) Stop() {
	phi.stop()
}

func (phi *phalanxImpl) PostTxs(txs []*commonProto.Transaction) {
	phi.postTxs(txs)
}

func (phi *phalanxImpl) Propose(comm *commonProto.CommMsg) {
	phi.propose(comm)
}