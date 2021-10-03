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

type QueryRequest struct {
	Threshold []uint64

	QueryList []QueryIndex
}
