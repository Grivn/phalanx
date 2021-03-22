package loggenerator

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/loggenerator/types"
)

func NewReliableLog(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) api.LogGenerator {
	return newLogGeneratorImpl(n, author, replyC, network, logger)
}

func (lg *logGeneratorImpl) Start() {
	lg.start()
}

func (lg *logGeneratorImpl) Stop() {
	lg.stop()
}

func (lg *logGeneratorImpl) Generate(bid *commonProto.BatchId) {
	lg.generate(bid)
}

func (lg *logGeneratorImpl) Record(msg *commonProto.OrderedMsg) {
	lg.record(msg)
}

func (lg *logGeneratorImpl) Ready(tag *commonProto.BinaryTag) {
	lg.ready(tag)
}
