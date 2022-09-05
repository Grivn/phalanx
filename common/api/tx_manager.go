package api

import "github.com/Grivn/phalanx/common/protos"

type Proposer interface {
	Runner
	TxProcessor
	SnappingUpTest
}

type TxProcessor interface {
	// ProcessTransaction is used to process transactions received by current node.
	ProcessTransaction(tx *protos.Transaction)
}

type SnappingUpTest interface {
	CommitResult(itemNo uint64, buyer uint64)
}
