package slot

import (
	"container/list"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type stableSlot struct {
	author uint64

	stableHeight []uint64

	pendingRequest *types.QueryRequest

	requestCache *list.List

	logger external.Logger
}

func (ss *stableSlot) executePending() {
	ss.pendingRequest = nil

	if ss.requestCache.Len() == 0 {
		return
	}

	front := ss.requestCache.Front()

	request, _ := front.Value.(*types.QueryRequest)

	ss.requestCache.Remove(front)

	ss.pendingRequest = request
}

func (ss *stableSlot) exist() bool {
	return ss.pendingRequest != nil
}

func (ss *stableSlot) postStableSlot(metaSlot []uint64) {
	request := ss.generateQueryRequest(metaSlot)

	if len(request.QueryList) == 0 {
		// we have not found any new sequence list to query.
		return
	}

	// if there isn't any pending request, just try it.
	if ss.pendingRequest == nil {
		ss.pendingRequest = request
		return
	}

	ss.requestCache.PushBack(request)
}

// generateQueryList is used to generate the query list for partial orders.
func (ss *stableSlot) generateQueryRequest(metaSlot []uint64) *types.QueryRequest {
	request := &types.QueryRequest{Threshold: metaSlot}

	for index, h := range ss.stableHeight {
		id := uint64(index+1)

		metaH := metaSlot[index]

		for i:=h+1; i<=metaH; i++ {
			qIdx := types.QueryIndex{Author: id, SeqNo: i}
			request.QueryList = append(request.QueryList, qIdx)
		}
	}

	return request
}
