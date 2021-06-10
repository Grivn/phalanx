package external

import "github.com/Grivn/phalanx/common/protos"

type ExecutorService interface {
	Execute(txs []*protos.Transaction, seqNo uint64, timestamp int64)
}
