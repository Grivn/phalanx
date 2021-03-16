package executor

import (
	"container/heap"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type logID struct {
	author uint64
	hash   string
}

type log struct {
	id    logID
	n     int
	f     int
	batch *commonProto.Batch
	heap  *timestampHeap
}

func newLog(n int,id logID) *log {
	return &log{
		id: id,
		n:  n,
		f:  (n-1)/4,
	}
}

func (l *log) len() int {
	return l.heap.Len()
}

func (l *log) quorum() int {
	return l.n - l.f
}

func (l *log) update(timestamp int64) {
	heap.Push(l.heap, timestamp)
}

func (l *log) assign(batch *commonProto.Batch) {
	l.batch = batch
}

func (l *log) isQuorum() bool {
	return l.len() >= l.quorum()
}

func (l *log) assigned() bool {
	return l.batch != nil
}

func (l *log) trustedTimestamp() int64 {
	return (*l.heap)[l.f]
}