package internal

import "github.com/Grivn/phalanx/common/protos"

type TxManager interface {
	// ProcessTransaction is used to process transactions received by current node.
	ProcessTransaction(tx *protos.Transaction)
}
