package types

import (
	"container/heap"
)

type TimestampHeap []int64

func NewTimestampHeap() TimestampHeapInterface {
	return &TimestampHeap{}
}

type TimestampHeapInterface interface {
	heap.Interface

	GetValue(which int) int64
}

func (h TimestampHeap) Len() int           { return len(h) }
func (h TimestampHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h TimestampHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *TimestampHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *TimestampHeap) Push(x interface{}) {
	*h = append(*h, x.(int64))
}

func (h *TimestampHeap) GetValue(which int) int64 {
	return (*h)[which-1]
}

type LogID struct {
	Author uint64
	Hash   string
}

type Log struct {
	ID   LogID
	N    int
	F    int
	Heap TimestampHeapInterface
}

func NewLog(n int,id LogID) *Log {
	return &Log{
		ID:   id,
		N:    n,
		F:    (n-1)/4,
		Heap: NewTimestampHeap(),
	}
}

func (l *Log) Len() int {
	return l.Heap.Len()
}

func (l *Log) Quorum() int {
	return l.N - l.F
}

func (l *Log) Update(timestamp int64) {
	heap.Push(l.Heap, timestamp)
}

func (l *Log) IsQuorum() bool {
	return l.Len() >= l.Quorum()
}

func (l *Log) TrustedTimestamp() int64 {
	return l.Heap.GetValue(l.F+1)
}