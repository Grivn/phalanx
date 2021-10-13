package types

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
)

// NewBlock generates the block to commit.
func NewBlock(command *protos.Command, timestamp int64) Block {
	return Block{Command: command, Timestamp: timestamp}
}

type Block struct {
	Command   *protos.Command
	Timestamp int64
}

func (block Block) Format() string {
	return fmt.Sprintf("[Block: command %s, trusted-timestamp %d]", block.Command.Format(), len(block.Command.Content))
}

// SortableBlocks is a slice of Block to sort.
type SortableBlocks []Block
func (s SortableBlocks) Len() int {
	return len(s)
}
func (s SortableBlocks) Less(i, j int) bool {
	if s[i].Timestamp == s[j].Timestamp {
		return s[i].Command.Digest < s[j].Command.Digest
	}
	return s[i].Timestamp < s[j].Timestamp
}
func (s SortableBlocks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
