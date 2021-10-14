package tracker

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type CommandTracker interface {
	RecordCommand(command *protos.Command)
	ReadCommand(digest string) *protos.Command
}

type PartialTracker interface {
	RecordPartial(pOrder *protos.PartialOrder)
	ReadPartial(idx types.QueryIndex) *protos.PartialOrder
}
