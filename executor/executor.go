package executor

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/executor/types"
	"github.com/Grivn/phalanx/external"
)

func NewExecutor(n int, author uint64, replyC chan types.ReplyEvent, executor external.Executor, logger external.Logger) api.Executor {
	return newExecuteImpl(n, author, replyC, executor, logger)
}

func (ei *executorImpl) Start() {
	ei.start()
}

func (ei *executorImpl) Stop() {
	ei.stop()
}

func (ei *executorImpl) ExecuteLogs(sequence uint64, logs []*commonProto.OrderedMsg){
	ei.executeLogs(sequence, logs)
}

func (ei *executorImpl) ExecuteBatch(batch *commonProto.Batch){
	ei.executeBatch(batch)
}
