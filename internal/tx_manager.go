package internal

import "github.com/Grivn/phalanx/common/protos"

type TxManager interface {
	ProcessTransaction(tx *protos.Transaction)
}
