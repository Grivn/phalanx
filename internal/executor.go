package internal

import "github.com/Grivn/phalanx/common/protos"

type Executor interface {
	// CommitQCs is used to commit the QCs.
	CommitQCs(qcb *protos.QCBatch) error
}
