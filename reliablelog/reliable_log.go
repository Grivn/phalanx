package reliablelog

import (
	"github.com/Grivn/phalanx/api"
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewReliableLog(n int, author uint64, sendC commonTypes.ReliableSendChan, network external.Network, logger external.Logger) api.ReliableLog {
	return newReliableLogImpl(n, author, sendC, network, logger)
}

func (rl *reliableLogImpl) Start() {
	rl.start()
}

func (rl *reliableLogImpl) Stop() {
	rl.stop()
}

func (rl *reliableLogImpl) Generate(bid *commonProto.BatchId) {
	rl.generate(bid)
}

func (rl *reliableLogImpl) RecordLog(log *commonProto.OrderedLog) {
	rl.recordLog(log)
}

func (rl *reliableLogImpl) RecordAck(ack *commonProto.OrderedAck) {
	rl.recordAck(ack)
}
