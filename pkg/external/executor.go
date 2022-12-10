package external

import (
	"github.com/Grivn/phalanx/pkg/common/types"
)

// Executor provides a service for block execution.
type Executor interface {
	// CommandExecution is used to execute a block.
	CommandExecution(block types.InnerBlock, seqNo uint64)
}
