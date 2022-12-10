package metrics

import (
	"github.com/Grivn/phalanx/pkg/common/types"
	"sync"
)

type SnappingUpMetrics struct {
	// mutex is used to process concurrency problem for this metrics instance.
	mutex sync.Mutex

	// committedItems
	committedItems map[uint64]bool

	// shoppingCarts is used to collect the number of items each buyer has bought.
	shoppingCarts map[uint64][]uint64
}

func NewSnappingUpMetrics() *SnappingUpMetrics {
	return &SnappingUpMetrics{
		committedItems: make(map[uint64]bool),
		shoppingCarts:  make(map[uint64][]uint64),
	}
}

func (m *SnappingUpMetrics) CommitSnappingUpResult(blk types.InnerBlock) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tick := blk.Command
	if tick == nil {
		return
	}

	itemNo := tick.Sequence
	buyer := tick.Author
	if m.committedItems[itemNo] {
		return
	}

	m.committedItems[itemNo] = true
	m.shoppingCarts[buyer] = append(m.shoppingCarts[buyer], itemNo)
}

func (m *SnappingUpMetrics) SuccessRates() []float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	maxID := uint64(0)
	sum := 0

	for id, cart := range m.shoppingCarts {
		if id > maxID {
			maxID = id
		}

		sum += len(cart)
	}

	res := make([]float64, int(maxID))
	for id, cart := range m.shoppingCarts {
		index := int(id - 1)
		res[index] = float64(len(cart)) / float64(sum)
	}
	return res
}
