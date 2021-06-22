package executor

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	// seqNo is used to track the sequence number for blocks.
	seqNo uint64

	// generator is used to generate blocks.
	generator *blockGenerator

	// exec is used to execute the block.
	exec external.ExecuteService
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(n int, exec external.ExecuteService) *executorImpl {
	return &executorImpl{generator: newBlockGenerator(n), exec: exec}
}

// CommitQCs is used to commit the QCs.
func (ei *executorImpl) CommitQCs(qcb *protos.QCBatch) error {
	sub, err := ei.generator.insertQCBatch(qcb)
	if err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	for _, blk := range sub {
		ei.seqNo++
		ei.exec.Execute(blk.CommandD, blk.TxList, ei.seqNo, blk.Timestamp)
	}

	return nil
}
