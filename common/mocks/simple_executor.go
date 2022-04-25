package mocks

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type executor struct {
	author uint64
	hash   string
	count  int
	logger external.Logger
}

func NewSimpleExecutor(author uint64, logger external.Logger) external.ExecutionService {
	return &executor{
		author: author,
		hash:   "initial",
		logger: logger,
	}
}

func (exe *executor) CommandExecution(block types.InnerBlock, seqNo uint64) {
	var list []string
	command := block.Command
	list = append(list, exe.hash)
	for _, tx := range command.Content {
		list = append(list, tx.Hash)
	}

	exe.count += len(command.Content)
	exe.hash = types.CalculateListHash(list, 0)
	if exe.author == uint64(1) {
		exe.logger.Infof("Author %d, FrontNo %d, Safe %v, Block Number %d, total len %d, Hash: %s, from Command %s",
			exe.author, block.FrontNo, block.Safe, seqNo, exe.count, exe.hash, command.Format())
	}
}
