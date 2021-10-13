package types

import (
	"fmt"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/google/btree"
)

type CommandIndex struct {
	Author uint64
	SeqNo  uint64
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
