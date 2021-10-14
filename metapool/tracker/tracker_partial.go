package tracker

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sync"
)

// partialTracker is used to record the partial orders current node has received.
// 1) receive partial order directly with the consensus process of logs.
// 2) fetch missing order process.
// the tracker of partial order belongs to sub instance, which means the partial orders from such a node.
type partialTracker struct {
	// mutex is used to control the concurrency problems of partial tracker.
	mutex sync.RWMutex

	// author indicates current node identifier.
	author uint64

	// partialMap records the partial orders which current node has received.
	partialMap map[types.QueryIndex]*protos.PartialOrder

	// logger prints logs.
	logger external.Logger
}

func NewPartialTracker(author uint64, logger external.Logger) *partialTracker {
	logger.Infof("[%d] initiate partial tracker")
	return &partialTracker{
		author:     author,
		partialMap: make(map[types.QueryIndex]*protos.PartialOrder),
		logger:     logger,
	}
}

func (pt *partialTracker) RecordPartial(pOrder *protos.PartialOrder) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	qIdx := types.QueryIndex{Author: pOrder.Author(), SeqNo: pOrder.Sequence()}

	if _, ok := pt.partialMap[qIdx]; ok {
		pt.logger.Debugf("[%d] duplicated partial order %s", pt.author, pOrder.Format())
		return
	}

	pt.partialMap[qIdx] = pOrder
}

func (pt *partialTracker) ReadPartial(idx types.QueryIndex) *protos.PartialOrder {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	// here, we are trying to read the partial order according to partial order query index.
	pOrder, ok := pt.partialMap[idx]
	if !ok {
		return nil
	}
	delete(pt.partialMap, idx)
	return pOrder
}
