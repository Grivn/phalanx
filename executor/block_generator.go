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

	executed map[string]bool
}

// newBlockGenerator is used to initiate an instance for block generation.
func newBlockGenerator(n int) *blockGenerator {
	return &blockGenerator{
		fault:   types.CalculateFault(n),
		quorum:  types.CalculateQuorum(n),
		pending: make(map[string]*types.PendingCommand),
		executed: make(map[string]bool),
	}
}

// insertBatch is used to insert the partial order into executor for block generation.
func (bg *blockGenerator) insertBatch(pBatch *protos.PartialOrderBatch) (types.SubBlock, error) {
	bg.mutex.Lock()
	defer bg.mutex.Unlock()

	// collect the commands which may be selected into the blocks.
	for digest, command := range pBatch.Commands {
		if _, ok := bg.pending[digest]; !ok {
			bg.pending[digest] = types.NewPendingCommand(command)
		}
	}

	// collect the partial order for block generation.
	for _, filter := range pBatch.PartialSet {
		for _, pOrder := range filter.PartialOrders {
			if bg.executed[pOrder.CommandDigest()] {
				continue
			}

			pc, ok := bg.pending[pOrder.CommandDigest()]
			if !ok {
				return nil, errors.New("invalid partial order")
			}

			pc.Replicas[pOrder.Author()] = true
			pc.Timestamps = append(pc.Timestamps, pOrder.Timestamp())
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
			CommandD:  pCommand.Command.Digest,
			TxList:    pCommand.Command.Content,
			HashList:  pCommand.Command.HashList,
			Timestamp: pCommand.Timestamps[bg.fault],
		}

		bg.executed[pCommand.Command.Digest] = true

		sub = append(sub, blk)
		delete(bg.pending, digest)
	}

	// sort the block according to timestamps
	sort.Sort(sub)
	return sub, nil
}
