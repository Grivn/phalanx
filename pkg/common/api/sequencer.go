package api

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
)

// Sequencer is used to generate commands and propose the order-attempt
type Sequencer interface {
	Runner

	// Sequencing is used to order the commands with pre-set strategies.
	Sequencing(command *protos.Command)
}

// SequencingEngine is used to cache the received commands and generate command_index to create order-attempts.
type SequencingEngine interface {
	// Sequencing is used to order the commands with pre-set strategies.
	Sequencing(command *protos.Command)

	// CommandIndexChan is used to produce the ordered command_index to generate order-attempts.
	CommandIndexChan() <-chan *types.CommandIndex
}

// Relay (module for experiment) is used to relay the commands from specific client with pre-defined ordering strategy.
type Relay interface {
	// Append is used to notify the latest received command from current client.
	Append(command *protos.Command) int
}
