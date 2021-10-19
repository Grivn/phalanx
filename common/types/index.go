package types

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"

	"github.com/google/btree"
)

//=============================================== Command Index ==========================================================

// CommandIndex is used to record the essential messages for commands we have received.
type CommandIndex struct {
	// Author indicates the generator of current command.
	Author uint64

	// SeqNo indicates command sequence number assigned by generator.
	SeqNo uint64

	// Digest indicates the identifier of current command.
	Digest string
}

func NewCommandIndex(command *protos.Command) *CommandIndex {
	return &CommandIndex{Author: command.Author, SeqNo: command.Sequence, Digest: command.Digest}
}

func (index *CommandIndex) Less(item btree.Item) bool {
	return index.SeqNo < (item.(*CommandIndex)).SeqNo
}

func (index *CommandIndex) Format() string {
	return fmt.Sprintf("[CommandIndex: client %d, sequence %d, digest %s]", index.Author, index.SeqNo, index.Digest)
}

//=============================================== Query Index ==========================================================

// QueryIndex indicates the query identifier for the scanner of partial orders.
type QueryIndex struct {
	// Author indicates the generator of one partial order.
	Author uint64

	// SeqNo indicates the partial order sequence assigned by generator.
	SeqNo uint64
}

func NewQueryIndex(author uint64, seqNo uint64) QueryIndex {
	return QueryIndex{Author: author, SeqNo: seqNo}
}

func (idx QueryIndex) Format() string {
	return fmt.Sprintf("[index: author %d, sequence %d]", idx.Author, idx.SeqNo)
}

// QueryStream is a series of query index for partial orders.
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