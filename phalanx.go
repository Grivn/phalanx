package phalanx

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type Phalanx interface {
	Start()

	Stop()

	IsNormal() bool

	PostTxs(txs []*commonProto.Transaction)

	Execute(payload []byte)
}
