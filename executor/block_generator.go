package executor

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type BlockGenerator interface {
	// InsertQCBatch is used to insert the QCs into executor for block generation.
	InsertQCBatch(qcb *protos.QCBatch) (types.SubBlock, error)
}
