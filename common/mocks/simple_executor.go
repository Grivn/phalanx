package mocks

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type executor struct {
	author uint64
	hash   string
	count  int
	logger external.Logger
}

func NewSimpleExecutor(author uint64, logger external.Logger) external.ExecuteService {
	return &executor{
		author: author,
		hash:   "initial",
		logger: logger,
	}
}

func (exe *executor) Execute(commandD string, txs []*protos.Transaction, seqNo uint64, timestamp int64) {
	var list []string

	list = append(list, exe.hash)
	for _, tx := range txs {
		list = append(list, tx.Hash)
	}

	exe.count += len(txs)
	exe.hash = types.CalculateListHash(list, timestamp)
	exe.logger.Infof("Author %d, Block Number %d, total len %d, Hash: %s, from Command %s", exe.author, seqNo, exe.count, exe.hash, commandD)
}
