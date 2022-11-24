package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

//================================== tracker for meta pool ========================================

// CommandTracker is used to record received commands.
type CommandTracker interface {
	RecordCommand(command *protos.Command)
	ReadCommand(digest string) *protos.Command
}

// PartialTracker is used to record received partial orders.
type PartialTracker interface {
	RecordPartial(pOrder *protos.PartialOrder)
	ReadPartial(idx types.QueryIndex) *protos.PartialOrder
	IsExist(idx types.QueryIndex) bool
}

// AttemptTracker is used to record received order-attempt related messages.
type AttemptTracker interface {
	// Record records order-attempt information.
	Record(attempt *protos.OrderAttempt)

	// Get gets order-attempt according to query index.
	Get(idx types.QueryIndex) *protos.OrderAttempt

	// Checkpoint records the checkpoint and take garbage-collection according to it.
	Checkpoint(checkpoint *protos.Checkpoint)

	// HasCheckpoint checks if current checkpoint exists.
	HasCheckpoint(idx types.QueryIndex) bool

	// GarbageCollect collects garbage according to current status.
	GarbageCollect()
}

type CheckpointTracker interface {
	Record(checkpoint *protos.Checkpoint)
	Get(idx types.QueryIndex) *protos.Checkpoint
	IsExist(idx types.QueryIndex) bool
}
