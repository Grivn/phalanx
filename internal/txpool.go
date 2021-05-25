package internal

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
)

type TxPool interface {
	Basic

	Reset()

	PostTx(tx *commonProto.Transaction)

	PostBatch(batch *commonProto.TxBatch)

	ExecuteBlock(block *commonTypes.Block)

	IsPoolFull() bool
}
