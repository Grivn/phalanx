package api

import "github.com/Grivn/phalanx/common/protos"

type Sequencer interface {
	Runner

	ReceiveTxs(transaction *protos.Transaction)
}

type PartialOrderStrategy interface {
}
