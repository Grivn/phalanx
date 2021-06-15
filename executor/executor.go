package executor

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type executorImpl struct {
	// seqNo is used to track the sequence number for blocks.
	seqNo uint64

	// generator is used to generate blocks.
	generator *blockGenerator

	// exec is used to execute the block.
	exec external.ExecutorService
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(n int, exec external.ExecutorService) *executorImpl {
	return &executorImpl{generator: newBlockGenerator(n), exec: exec}
}

// CommitQCs is used to commit the QCs.
func (ei *executorImpl) CommitQCs(payload []byte) error {
	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	sub, err := ei.generator.insertQCBatch(qcb)
	if err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	for _, blk := range sub {
		ei.seqNo++
		ei.exec.Execute(blk.TxList, ei.seqNo, blk.Timestamp)
	}

	return nil
}
