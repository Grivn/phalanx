package types

import "github.com/Grivn/phalanx/common/protos"

type Timestamps []int64

func (t Timestamps) Len() int           { return len(t) }
func (t Timestamps) Less(i, j int) bool { return t[i] < t[j] }
func (t Timestamps) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

type PendingCommand struct {
	Replicas   map[uint64]bool
	Command    *protos.Command
	Timestamps Timestamps
}

type Block struct {
	TxList    []*protos.Transaction
	HashList  []string
	Timestamp int64
}

type SubBlock []Block

func (s SubBlock) Len() int           { return len(s) }
func (s SubBlock) Less(i, j int) bool { return s[i].Timestamp < s[j].Timestamp }
func (s SubBlock) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func NewPendingCommand(command *protos.Command) *PendingCommand {
	return &PendingCommand{Replicas: make(map[uint64]bool), Command: command, Timestamps: nil}
}

