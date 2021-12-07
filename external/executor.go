package external

import (
	"github.com/Grivn/phalanx/common/types"
)

// ExecutionService provides a service for block execution.
type ExecutionService interface {
	// CommandExecution is used to execute a block.
	CommandExecution(block types.InnerBlock, seqNo uint64)
}
