package internal

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
)

type TxManager interface {
	api.Runner
	// ProcessTransaction is used to process transactions received by current node.
	ProcessTransaction(tx *protos.Transaction)
}
