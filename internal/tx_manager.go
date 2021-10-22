package internal

import "github.com/Grivn/phalanx/common/protos"

type TxManager interface {
	// Run starts tx manager coroutine service.
	Run()

	// Close stops tx manager coroutine service.
	Close()

	// ProcessTransaction is used to process transactions received by current node.
	ProcessTransaction(tx *protos.Transaction)
}
