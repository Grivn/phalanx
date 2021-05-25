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

func (rl *reliableLogImpl) Generate(bid *commonProto.BatchId) {
	rl.generate(bid)
}

func (rl *reliableLogImpl) RecordLog(log *commonProto.OrderedLog) {
	rl.recordLog(log)
}

func (rl *reliableLogImpl) RecordAck(ack *commonProto.OrderedAck) {
	rl.recordAck(ack)
}
