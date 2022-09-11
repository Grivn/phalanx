package api

import "github.com/Grivn/phalanx/common/protos"

type Proposer interface {
	Runner

	// ProcessTransaction is used to process transactions received by current node.
	ProcessTransaction(tx *protos.Transaction)
}
