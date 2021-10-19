package tracker

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

// CommandTracker is used to record received commands.
type CommandTracker interface {
	RecordCommand(command *protos.Command)
	ReadCommand(digest string) *protos.Command
}

// PartialTracker is used to record received partial orders.
type PartialTracker interface {
	RecordPartial(pOrder *protos.PartialOrder)
	ReadPartial(idx types.QueryIndex) *protos.PartialOrder
}
