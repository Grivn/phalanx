package reliablelog

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/reliablelog/types"
)

func NewReliableLog(n int, author uint64, replyC chan types.ReplyEvent, auth api.Authenticator, network external.Network, logger external.Logger) api.ReliableLog {
	return newReliableLogImpl(n, author, replyC, auth, network, logger)
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

func (rl *reliableLogImpl) Record(msg *commonProto.SignedMsg) {
	rl.record(msg)
}

func (rl *reliableLogImpl) Ready(tag *commonProto.BinaryTag) {
	rl.ready(tag)
}
