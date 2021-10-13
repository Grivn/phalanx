package mocks

import (
	"fmt"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type executor struct {
	author uint64
	hash   string
	count  int

	commandVerifier map[uint64]uint64

	logger external.Logger
}

func NewSimpleExecutor(author uint64, logger external.Logger) external.ExecutionService {
	return &executor{
		author: author,
		hash:   "initial",
		logger: logger,
		commandVerifier: make(map[uint64]uint64),
	}
}

func (exe *executor) CommandExecution(command *protos.Command, seqNo uint64, timestamp int64) {
	var list []string

	if command.Sequence != exe.commandVerifier[command.Author]+1 {
		panic(fmt.Sprintf("invalid command sequence from %d, expect %d, received %d", command.Author, exe.commandVerifier[command.Author]+1, command.Sequence))
	}

	list = append(list, exe.hash)
	for _, tx := range command.Content {
		list = append(list, tx.Hash)
	}

	exe.commandVerifier[command.Author]++

	exe.count += len(command.Content)
	exe.hash = types.CalculateListHash(list, 0)
	exe.logger.Infof("Author %d, Block Number %d, total len %d, Hash: %s, from Command %s", exe.author, seqNo, exe.count, exe.hash, command.Digest)
	//exe.logger.Debugf("Execution list %v", exe.commandVerifier)
}
