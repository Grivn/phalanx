package binsubset

import (
	"github.com/Grivn/phalanx/api"
	"github.com/Grivn/phalanx/binsubset/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

func NewSubset(n int, author uint64, replyC chan types.ReplyEvent, network external.Network, logger external.Logger) api.BinaryByzantine {
	return newSubsetImpl(n, author, replyC, network, logger)
}

func (si *subsetImpl) Start() {
	si.start()
}

func (si *subsetImpl) Stop() {
	si.stop()
}

func (si *subsetImpl) Trigger(tag *commonProto.BinaryTag) {
	si.trigger(tag)
}

func (si *subsetImpl) Propose(ntf *commonProto.BinaryNotification) {
	si.propose(ntf)
}
