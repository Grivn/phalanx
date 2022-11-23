package api

import "github.com/Grivn/phalanx/common/protos"

type Sequencer interface {
	Runner
	MetaCommitter

	// ProcessCommand is used to receive commands and generate order-attempt with them.
	ProcessCommand(command *protos.Command)
}
