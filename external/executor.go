package external

import "github.com/Grivn/phalanx/common/protos"

type Executor interface {
	Execute(txs []*protos.Transaction, localList []bool, seqNo uint64, timestamp int64)
}
