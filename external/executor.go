package external

import "github.com/Grivn/phalanx/common/protos"

// ExecutionService provides a service for block execution.
type ExecutionService interface {
	// CommandExecution is used to execute a block.
	CommandExecution(command *protos.Command, seqNo uint64, timestamp int64)
}
