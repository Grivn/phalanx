package executor

type timestampHeap []int64

func (h timestampHeap) Len() int           { return len(h) }
func (h timestampHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h timestampHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *timestampHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *timestampHeap) Push(x interface{}) {
	*h = append(*h, x.(int64))
}
