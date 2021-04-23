package reliablelog

import (
	"time"

	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type generator struct {
	author   uint64
	sequence uint64
	sender   external.Network
	logger   external.Logger
}

func newGenerator(author uint64, network external.Network, logger external.Logger) *generator {
	return &generator{
		author:   author,
		sequence: uint64(0),
		sender:   network,
		logger:   logger,
	}
}

func (generator *generator) generate(bid *commonProto.BatchId) *commonProto.OrderedLog {
	generator.sequence++

	log := &commonProto.OrderedLog{
		Author:    generator.author,
		Sequence:  generator.sequence,
		BatchId:   bid,
		Timestamp: time.Now().UnixNano(),
	}

	generator.logger.Infof("[GENERATE LOG] replica %d sequence %d", generator.author, generator.sequence)
	generator.sender.BroadcastLog(log)
	return log
}
