package streamcache

import (
	"container/list"
	"sync"

	"github.com/Grivn/phalanx/pkg/common/api"
)

type streamCache struct {
	// mutex is used to process the concurrency of streams processing.
	mutex sync.Mutex

	// streamList is used to record the committed query stream.
	streamList *list.List
}

func NewStreamCache() api.StreamCache {
	return &streamCache{
		streamList: list.New(),
	}
}

func (mgr *streamCache) Append(item interface{}) {
	// append the query stream into stream list.
	mgr.mutex.Lock()
	mgr.streamList.PushBack(item)
	mgr.mutex.Unlock()
}

func (mgr *streamCache) Front() interface{} {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if mgr.streamList.Len() == 0 {
		// no values in stream list, return nil.
		return nil
	}

	// pop the first value in the stream list.
	item := mgr.streamList.Front()
	mgr.streamList.Remove(item)
	return item.Value
}