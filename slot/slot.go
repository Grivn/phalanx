package slot

import (
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"sync"
)

// update the slot every time we received quorum cert for one node.

// if we have received enough partial slot, we would pick out the partial orders which we need to g

type slotManager struct {
	mutex sync.Mutex

	partial *partialSlot

	stable *stableSlot

	sp internal.SequencePool

	exec internal.Execution

	logger external.Logger
}

func (slot *slotManager) PartialHeight(author, height uint64) {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	slot.partial.updatePartialSlot(author, height)
	slot.tryExecution()
}

func (slot *slotManager) StableSlot(metaSlot []uint64) {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	slot.stable.postStableSlot(metaSlot)
	slot.tryExecution()
}

func (slot *slotManager) tryExecution() {
	if slot.stable.exist() {
		stableS := slot.stable.pendingRequest

		if !slot.partial.verifyRemoteSlot(stableS.Threshold) {
			return
		}

		slot.stable.executePending()

		// todo query the partial orders from sequence pool and execute them with execution module.

		slot.tryExecution()
	}
}
