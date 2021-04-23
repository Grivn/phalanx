package executor

import (
	"github.com/Grivn/phalanx/api"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewExecutor(n int, author uint64, sendC commonTypes.ExecutorSendChan, logger external.Logger) api.Executor {
	return newExecuteImpl(n, author, sendC, logger)
}

func (ei *executorImpl) Start() {
	ei.start()
}

func (ei *executorImpl) Stop() {
	ei.stop()
}

func (ei *executorImpl) ExecuteLogs(exec *commonProto.ExecuteLogs){
	ei.executeLogs(exec)
}
