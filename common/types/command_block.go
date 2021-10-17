package types

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
)

// NewInnerBlock generates the phalanx inner block to commit.
func NewInnerBlock(command *protos.Command, timestamp int64) InnerBlock {
	return InnerBlock{Command: command, Timestamp: timestamp}
}

// InnerBlock is an executed block for phalanx.
type InnerBlock struct {
	// Command is the content of current block.
	Command *protos.Command

	// Timestamp is the trusted time for current block generation.
	Timestamp int64
}

func (block InnerBlock) Format() string {
	return fmt.Sprintf("[Block: command %s, trusted-timestamp %d]", block.Command.Format(), len(block.Command.Content))
}

// SortableInnerBlocks is a slice of inner block to sort.
type SortableInnerBlocks []InnerBlock
func (s SortableInnerBlocks) Len() int {
	return len(s)
}
func (s SortableInnerBlocks) Less(i, j int) bool {
	if s[i].Timestamp == s[j].Timestamp {
		return s[i].Command.Digest < s[j].Command.Digest
	}
	return s[i].Timestamp < s[j].Timestamp
}
func (s SortableInnerBlocks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
