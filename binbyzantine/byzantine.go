package binbyzantine

import (
	"github.com/Grivn/phalanx/api"
	"github.com/Grivn/phalanx/binbyzantine/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewByzantine(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) api.BinaryByzantine {
	return newByzantineImpl(n, author, replyC, network, logger)
}

func (bi *byzantineImpl) Start() {
	bi.start()
}

func (bi *byzantineImpl) Stop() {
	bi.stop()
}

func (bi *byzantineImpl) Trigger(tag *commonProto.BinaryTag) {
	bi.trigger(tag)
}

func (bi *byzantineImpl) Propose(ntf *commonProto.BinaryNotification) {
	bi.propose(ntf)
}
