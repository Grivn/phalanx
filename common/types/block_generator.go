package types

import "github.com/Grivn/phalanx/common/protos"

// Timestamps is a slice of timestamps to sort.
type Timestamps []int64

func (t Timestamps) Len() int           { return len(t) }
func (t Timestamps) Less(i, j int) bool { return t[i] < t[j] }
func (t Timestamps) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

type PendingCommand struct {
	Replicas   map[uint64]bool
	Command    *protos.Command
	Timestamps Timestamps
}

// NewPendingCommand is used to generate PendingCommand struct for executor.
func NewPendingCommand(command *protos.Command) *PendingCommand {
	return &PendingCommand{Replicas: make(map[uint64]bool), Command: command, Timestamps: nil}
}

func NewBlock(commandD string, txList []*protos.Transaction, hashList []string, timestamp int64) Block {
	return Block{CommandD: commandD, TxList: txList, HashList: hashList, Timestamp: timestamp}
}

type Block struct {
	CommandD  string
	TxList    []*protos.Transaction
	HashList  []string
	Timestamp int64
}

// SubBlock is a slice of Block to sort.
type SubBlock []Block
func (s SubBlock) Len() int {
	return len(s)
}
func (s SubBlock) Less(i, j int) bool {
	if s[i].Timestamp == s[j].Timestamp {
		return s[i].CommandD < s[j].CommandD
	}
	return s[i].Timestamp < s[j].Timestamp
}
func (s SubBlock) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
