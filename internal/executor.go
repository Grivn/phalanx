package internal

import "github.com/Grivn/phalanx/common/protos"

type Executor interface {
	// CommitPartials is used to commit the partial orders.
	CommitPartials(qcb *protos.PartialOrderBatch) error
}

type Execution interface {
	CommitPartials(pOrders []*protos.PartialOrder)
}
