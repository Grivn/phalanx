package phalanx

import commonProto "github.com/Grivn/phalanx/common/protos"

type Phalanx interface {
	Start()

	Stop()

	IsNormal() bool

	PostTxs(txs []*commonProto.Transaction)

	Execute(payload []byte)
}
