package external

import "github.com/Grivn/phalanx/common/protos"

// ExecuteService provides a service for block execution.
type ExecuteService interface {
	// Execute is used to execute a block.
	Execute(commandD string, txs []*protos.Transaction, seqNo uint64, timestamp int64)
}
