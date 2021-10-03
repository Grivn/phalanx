package slot

import "github.com/Grivn/phalanx/external"

type partialSlot struct {
	author uint64

	N int

	// the index for slot height is the identifier for each participant.

	slotHeight []uint64

	logger external.Logger
}

func (ps *partialSlot) updatePartialSlot(author, height uint64) {

	index := int(author-1)

	// check the height for partial slot height.
	if ps.slotHeight[index] < height {
		ps.logger.Debugf("[%d] expired height, current %d, update %d", ps.author, ps.slotHeight[author], height)
		return
	}

	// update the height for the partial chain.
	ps.slotHeight[index] = height
}

// verifyRemoteSlot is used to verify the remote slot valid for consensus.
func (ps *partialSlot) verifyRemoteSlot(remoteSlot []uint64) bool {
	for index, remoteHeight := range remoteSlot {
		if ps.slotHeight[index] < remoteHeight {
			return false
		}
	}
	return true
}

// collectPartialSlot returns the height we have known for each node.
func (ps *partialSlot) collectPartialSlot() []uint64 {
	ps.logger.Debugf("[%d] collect partial slots %v", ps.slotHeight)
	return ps.slotHeight
}
