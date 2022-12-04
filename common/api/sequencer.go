package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

// Sequencer is used to generate commands and propose the order-attempt
type Sequencer interface {
	Runner

	ReceiveLocalEvent(event types.LocalEvent)
}

// SequencingEngine is used to cache the received commands and generate command_index to create order-attempts.
type SequencingEngine interface {
	Sequencing(command *protos.Command)
}

// Relay (module for experiment) is used to relay the commands from specific client with pre-defined ordering strategy.
type Relay interface {
	// Append is used to notify the latest received command from current client.
	Append(command *protos.Command) int
}
