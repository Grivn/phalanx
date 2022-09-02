package tracker

import (
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

// partialTracker is used to record the partial orders current node has received.
// 1) receive partial order directly with the consensus process of logs.
// 2) fetch missing order process.
// the tracker of partial order belongs to sub instance, which means the partial orders from such a node.
type partialTracker struct {
	// author indicates current node identifier.
	author uint64

	// partialMap records the partial orders which current node has received.
	partialMap sync.Map

	// logger prints logs.
	logger external.Logger
}

func NewPartialTracker(author uint64, logger external.Logger) api.PartialTracker {
	logger.Infof("[%d] initiate partial tracker")
	return &partialTracker{
		author: author,
		logger: logger,
	}
}

func (pt *partialTracker) RecordPartial(pOrder *protos.PartialOrder) {
	qIdx := types.QueryIndex{Author: pOrder.Author(), SeqNo: pOrder.Sequence()}

	if _, ok := pt.partialMap.Load(qIdx); ok {
		pt.logger.Debugf("[%d] duplicated partial order %s", pt.author, pOrder.Format())
		return
	}

	pt.partialMap.Store(qIdx, pOrder)
}

func (pt *partialTracker) ReadPartial(idx types.QueryIndex) *protos.PartialOrder {
	// here, we are trying to read the partial order according to partial order query index.
	e, ok := pt.partialMap.Load(idx)
	if !ok {
		return nil
	}
	pOrder := e.(*protos.PartialOrder)
	pt.partialMap.Delete(idx)
	return pOrder
}

func (pt *partialTracker) IsExist(idx types.QueryIndex) bool {
	_, ok := pt.partialMap.Load(idx)
	return ok
}
