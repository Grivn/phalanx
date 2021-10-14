package instance

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
)

type ClientInstance interface {
	api.Runner
	Commit(seqNo uint64)
	Append(command *protos.Command)
}

type ReplicaInstance interface {
	GetHighOrder() *protos.PartialOrder
	ReceivePreOrder(pre *protos.PreOrder) error
	ReceivePartial(pOrder *protos.PartialOrder) error
}
