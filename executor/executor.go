package executor

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

func NewExecutor(n int, author uint64, sendC commonTypes.ExecutorSendChan, logger external.Logger) internal.Executor {
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
