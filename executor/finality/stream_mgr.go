package finality

import (
	"container/list"
	"github.com/Grivn/phalanx/common/types"
	"sync"
)

type streamCache struct {
	// mutex is used to process the concurrency of streams processing.
	mutex sync.Mutex

	// streamList is used to record the committed query stream.
	streamList *list.List
}

func newStreamCache() *streamCache {
	return &streamCache{
		streamList: list.New(),
	}
}

func (mgr *streamCache) append(qStream types.QueryStream) {
	//mgr.mutex.Lock()
	//defer mgr.mutex.Unlock()

	if len(qStream) == 0 {
		return
	}

	// append the query stream into stream list.
	mgr.streamList.PushBack(qStream)
}

func (mgr *streamCache) front() types.QueryStream {
	//mgr.mutex.Lock()
	//defer mgr.mutex.Unlock()

	if mgr.streamList.Len() == 0 {
		return nil
	}
	return mgr.streamList.Front().Value.(types.QueryStream)
}
