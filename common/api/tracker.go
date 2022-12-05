package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

//================================== tracker for meta pool ========================================

// CommandTracker is used to record received commands.
type CommandTracker interface {
	Record(command *protos.Command)
	Get(digest string) *protos.Command
}

// PartialTracker is used to record received partial orders.
type PartialTracker interface {
	Record(pOrder *protos.PartialOrder)
	Get(idx types.QueryIndex) *protos.PartialOrder
	IsExist(idx types.QueryIndex) bool
}

// AttemptTracker is used to record received order-attempts.
type AttemptTracker interface {
	Record(attempt *protos.OrderAttempt)
	Get(idx types.QueryIndex) *protos.OrderAttempt
}

// CheckpointTracker is used to record received checkpoint for order-attempts.
type CheckpointTracker interface {
	Record(checkpoint *protos.Checkpoint)
	Get(idx types.QueryIndex) *protos.Checkpoint
	IsExist(idx types.QueryIndex) bool
}
