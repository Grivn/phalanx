package api

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type TxPool interface {
	Basic

	Reset()

	PostTx(tx *commonProto.Transaction)

	PostBatch(batch *commonProto.Batch)

	Load(bid *commonProto.BatchId)
}
