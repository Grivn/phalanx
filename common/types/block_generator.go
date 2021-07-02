package types

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
)

// NewBlock generates the block to commit.
func NewBlock(commandD string, txList []*protos.Transaction, hashList []string, timestamp int64) Block {
	return Block{CommandD: commandD, TxList: txList, HashList: hashList, Timestamp: timestamp}
}

type Block struct {
	CommandD  string
	TxList    []*protos.Transaction
	HashList  []string
	Timestamp int64
}

func (block Block) Format() string {
	return fmt.Sprintf("<digest %s, txCoount %d>", block.CommandD, len(block.TxList))
}

// SortableBlocks is a slice of Block to sort.
type SortableBlocks []Block
func (s SortableBlocks) Len() int {
	return len(s)
}
func (s SortableBlocks) Less(i, j int) bool {
	if s[i].Timestamp == s[j].Timestamp {
		return s[i].CommandD < s[j].CommandD
	}
	return s[i].Timestamp < s[j].Timestamp
}
func (s SortableBlocks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
