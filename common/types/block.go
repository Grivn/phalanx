package types

import "container/heap"

type LogHeap struct {
	Heap     []*Log
	Presence map[LogID]*Log
}

type LogHeapInterface interface {
	heap.Interface

	GetSlice() []*Log
}

func NewLogHeap() LogHeapInterface {
	return &LogHeap{}
}

func (l LogHeap) Len() int      { return len(l.Heap) }
func (l LogHeap) Less(i, j int) bool { return l.Heap[i].TrustedTimestamp() < l.Heap[j].TrustedTimestamp() }
func (l LogHeap) Swap(i, j int) { l.Heap[i], l.Heap[j] = l.Heap[j], l.Heap[i] }

func (l *LogHeap) Pop() interface{} {
	n := len(l.Heap)
	x := l.Heap[n-1]
	l.Heap = l.Heap[0 : n-1]
	return x
}

func (l *LogHeap) Push(x interface{}) {
	val := x.(*Log)
	_, ok := l.Presence[val.ID]
	if ok {
		return
	}
	l.Heap = append(l.Heap, x.(*Log))
}

func (l *LogHeap) GetSlice() []*Log {
	return l.Heap
}

type Block struct {
	Sequence  uint64
	Logs      []*Log
	Pending   map[LogID]bool
	Timestamp int64
}

func NewBlock(sequence uint64, logs []*Log) *Block {
	assigned := make(map[LogID]bool)
	for _, log := range logs {
		assigned[log.ID] = true
	}

	return &Block{
		Sequence:  sequence,
		Logs:      logs,
		Pending:   assigned,
		Timestamp: logs[0].TrustedTimestamp(),
	}
}
