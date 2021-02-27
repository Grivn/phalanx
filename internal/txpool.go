package internal

import (
	commonBasic "github.com/Grivn/phalanx/common/basic"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type TxPool interface {
	commonBasic.Basic

	PostTx(tx *commonProto.Transaction)

	PostBatch(batch *commonProto.Batch)

	Load(bid *commonProto.BatchId)
}
