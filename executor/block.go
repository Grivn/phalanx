package executor

import "sort"

type LogSlice []*log

func (l LogSlice) Len() int           { return len(l) }
func (l LogSlice) Less(i, j int) bool { return l[i].trustedTimestamp() < l[j].trustedTimestamp() }
func (l LogSlice) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

type block struct {
	logs LogSlice

	count int

	timestamp int64
}

func newBlock(logs []*log) *block {
	raw := LogSlice(logs)

	sort.Sort(raw)

	return &block{
		logs:      raw,
		count:     len(raw),
		timestamp: logs[0].trustedTimestamp(),
	}
}

func (block *block) logSlice() []*log {
	return block.logs
}
