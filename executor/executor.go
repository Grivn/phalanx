package executor

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/executor/blockgenerator"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type executorImpl struct {
	seq uint64

	bg BlockGenerator

	// exec is used to execute the block.
	exec external.ExecutorService
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(n int, exec external.ExecutorService) *executorImpl {
	return &executorImpl{bg: blockgenerator.NewBlockGenerator(n), exec: exec}
}

// CommitQCs is used to commit the QCs.
func (ei *executorImpl) CommitQCs(payload []byte) error {
	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	sub, err := ei.bg.InsertQCBatch(qcb)
	if err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	fmt.Println(len(sub))

	for _, blk := range sub {
		ei.seq++
		ei.exec.Execute(blk.TxList, ei.seq, blk.Timestamp)
	}

	return nil
}
