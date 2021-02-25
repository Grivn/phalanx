package external

import (
	commonBasic "github.com/Grivn/phalanx/common/basic"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type TxPool interface {
	commonBasic.Basic

	PostTxs(tx *commonProto.Transaction)

	LoadTxs(list []string) []*commonProto.Transaction
}
