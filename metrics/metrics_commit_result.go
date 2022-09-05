package metrics

type CommitResultMetrics struct {
	// committedItems
	committedItems map[uint64]bool

	// shoppingCarts is used to collect the number of items each buyer has bought.
	shoppingCarts map[uint64][]uint64
}

func NewCommitResultMetrics() *CommitResultMetrics {
	return &CommitResultMetrics{
		committedItems: make(map[uint64]bool),
		shoppingCarts:  make(map[uint64][]uint64),
	}
}

func (cm *CommitResultMetrics) CommitSnappingUpResult(itemNo uint64, buyer uint64) {
	mutex.Lock()
	defer mutex.Unlock()

	if cm.committedItems[itemNo] {
		return
	}

	cm.committedItems[itemNo] = true
	cm.shoppingCarts[buyer] = append(cm.shoppingCarts[buyer], itemNo)
}

func (cm *CommitResultMetrics) Proportion() []float64 {
	mutex.Lock()
	defer mutex.Unlock()

	maxID := uint64(0)
	sum := 0

	for id, cart := range cm.shoppingCarts {
		if id > maxID {
			maxID = id
		}

		sum += len(cart)
	}

	res := make([]float64, int(maxID))
	for id, cart := range cm.shoppingCarts {
		index := int(id - 1)
		res[index] = float64(len(cart)) / float64(sum)
	}
	return res
}
