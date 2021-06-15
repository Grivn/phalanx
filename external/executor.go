package external

import "github.com/Grivn/phalanx/common/protos"

type ExecutorService interface {
	// Execute is used to execute a block.
	Execute(txs []*protos.Transaction, seqNo uint64, timestamp int64)
}
