package types

type Block struct {
	Sequence  uint64
	Logs      []*ExecuteLog
	Timestamp int64
}

func (blk *Block) Len() int           { return len(blk.Logs) }
func (blk *Block) Less(i, j int) bool { return blk.Logs[i].TrustedTimestamp() < blk.Logs[j].TrustedTimestamp() }
func (blk *Block) Swap(i, j int)      { blk.Logs[i], blk.Logs[j] = blk.Logs[j], blk.Logs[i] }

func NewBlock(sequence uint64, logs []*ExecuteLog) *Block {
	return &Block{
		Sequence:  sequence,
		Logs:      logs,
	}
}

func (blk *Block) UpdateTimestamp() {
	blk.Timestamp = blk.Logs[0].TrustedTs
}
