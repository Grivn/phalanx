package api

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type TxPool interface {
	Basic

	Reset()

	PostTx(tx *commonProto.Transaction)

	PostBatch(batch *commonProto.Batch)

	ExecuteBlock(block *commonTypes.Block)

	IsPoolFull() bool
}
