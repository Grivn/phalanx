package slot

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type partialTracker struct {
	// author indicates current node identifier.
	author uint64

	// partialMap records the partial orders which current node has received.
	partialMap map[types.QueryIndex]*protos.PartialOrder

	// logger prints logs.
	logger external.Logger
}

func newPartialTracker(author uint64, logger external.Logger) *partialTracker {
	logger.Infof("[%d] initiate partial tracker")
	return &partialTracker{
		author:     author,
		partialMap: make(map[types.QueryIndex]*protos.PartialOrder),
		logger:     logger,
	}
}

func (pt *partialTracker) recordPartial(pOrder *protos.PartialOrder) {
	qIdx := types.QueryIndex{Author: pOrder.Author(), SeqNo: pOrder.Sequence()}

	if _, ok := pt.partialMap[qIdx]; ok {
		pt.logger.Debugf("[%d] duplicated partial order %s", pt.author, pOrder.Format())
		return
	}

	pt.partialMap[qIdx] = pOrder
}

func (pt *partialTracker) readPartial(idx types.QueryIndex) *protos.PartialOrder {
	pOrder, ok := pt.partialMap[idx]
	if !ok {
		return nil
	}
	return pOrder
}
