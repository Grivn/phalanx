package types

import "fmt"

// QueryIndex indicates the query identifier for the scanner of partial orders.
type QueryIndex struct {
	Author uint64
	SeqNo  uint64
}

func (idx QueryIndex) Format() string {
	return fmt.Sprintf("[index: author %d, sequence %d]", idx.Author, idx.SeqNo)
}

type QueryStream []QueryIndex

func (qs QueryStream) Len() int { return len(qs) }
func (qs QueryStream) Swap(i, j int) { qs[i], qs[j] = qs[j], qs[i] }
func (qs QueryStream) Less(i, j int) bool {
	// query index for the same node, sort according to sequence number.
	if qs[i].SeqNo == qs[j].SeqNo {
		return qs[i].Author < qs[j].Author
	}

	// sort according to author.
	return qs[i].SeqNo < qs[j].SeqNo
}
