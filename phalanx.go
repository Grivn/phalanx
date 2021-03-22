package phalanx

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type Phalanx interface {
	Start()

	Stop()

	PostTxs(txs []*commonProto.Transaction)

	Propose(comm *commonProto.CommMsg)

	IsNormal() bool
}
