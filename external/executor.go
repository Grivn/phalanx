package external

import "github.com/Grivn/phalanx/common/protos"

// ExecutionService provides a service for block execution.
type ExecutionService interface {
	// CommandExecution is used to execute a block.
	CommandExecution(commandD string, txs []*protos.Transaction, seqNo uint64, timestamp int64)
}
