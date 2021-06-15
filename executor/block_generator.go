package executor

import (
	"errors"
	"github.com/Grivn/phalanx/common/types"
	"sort"
	"sync"

	"github.com/Grivn/phalanx/common/protos"
)

// blockGenerator is used to generate a real block according to the consensus result of bft.
type blockGenerator struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.Mutex

	// fault indicates the upper limit for byzantine nodes.
	fault int

	// quorum indicates the legal size for bft.
	quorum int

	// pending is used to track the commands which have been verified by bft consensus for execution.
	pending map[string]*types.PendingCommand
}

// newBlockGenerator is used to initiate an instance for block generation.
func newBlockGenerator(n int) *blockGenerator {
	return &blockGenerator{
		fault:   types.CalculateFault(n),
		quorum:  types.CalculateQuorum(n),
		pending: make(map[string]*types.PendingCommand),
	}
}

// insertQCBatch is used to insert the QCs into executor for block generation.
func (bg *blockGenerator) insertQCBatch(qcb *protos.QCBatch) (types.SubBlock, error) {
	bg.mutex.Lock()
	defer bg.mutex.Unlock()

	// collect the commands which may be selected into the blocks.
	for digest, command := range qcb.Commands {
		if _, ok := bg.pending[digest]; !ok {
			bg.pending[digest] = types.NewPendingCommand(command)
		}
	}

	// collect the QCs for block generation.
	for _, filter := range qcb.Filters {
		for _, qc := range filter.QCs {
			pc, ok := bg.pending[qc.CommandDigest()]
			if !ok {
				return nil, errors.New("invalid QC")
			}

			pc.Replicas[qc.Author()] = true
			pc.Timestamps = append(pc.Timestamps, qc.Timestamp())
		}
	}

	var sub types.SubBlock
	for digest, pCommand := range bg.pending {
		// select the commands which have reached the quorum size for block generation.
		if len(pCommand.Replicas) < bg.quorum {
			continue
		}

		// construct the block entity
		sort.Sort(pCommand.Timestamps)
		blk := types.Block{
			TxList:    pCommand.Command.Content,
			HashList:  pCommand.Command.HashList,
			Timestamp: pCommand.Timestamps[bg.fault],
		}

		sub = append(sub, blk)
		delete(bg.pending, digest)
	}

	// sort the block according to timestamps
	sort.Sort(sub)
	return sub, nil
}
