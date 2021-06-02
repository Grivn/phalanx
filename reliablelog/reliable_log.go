package reliablelog

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

func NewReliableLog(n int, author uint64, sendC commonTypes.ReliableSendChan, network external.Network, logger external.Logger) internal.ReliableLog {
	return newReliableLogImpl(n, author, sendC, network, logger)
}

func (rl *reliableLogImpl) Start() {
	rl.start()
}

func (rl *reliableLogImpl) Stop() {
	rl.stop()
}

func (rl *reliableLogImpl) ProcessCommand(command *commonProto.Command) {
	pre, err := rl.generator.generatePreOrder(command)
	if err != nil {
		rl.logger.Errorf("%s", err)
		return
	}


}

func (rl *reliableLogImpl) ProcessPreOrder(pre *commonProto.PreOrder) {
	err := rl.subInstances[pre.Author].ProcessPreOrder(pre)
	if err != nil {
		rl.logger.Errorf("%s", err)
	}
}

func (rl *reliableLogImpl) ProcessVote(vote *commonProto.Vote) {}

func (rl *reliableLogImpl) ProcessOrder(order *commonProto.Order) {}

func (rl *reliableLogImpl) Generate(bid *commonProto.BatchId) {}

func (rl *reliableLogImpl) RecordLog(log *commonProto.OrderedLog) {}

func (rl *reliableLogImpl) RecordAck(ack *commonProto.OrderedAck) {}
