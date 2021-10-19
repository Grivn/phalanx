package executor

import "github.com/Grivn/phalanx/common/protos"

type CommandRecorder interface {
	StoreCommand(command *protos.Command)

	ReadCommandRaw(commandD string) *protos.Command
	ReadCommandInfo(commandD string) *CommandInfo

	ReadCSCInfos() []*CommandInfo
}
