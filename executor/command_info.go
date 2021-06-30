package executor

import "github.com/Grivn/phalanx/common/protos"

// sortableTimestamps is a sortable slice for free will trusted timestamp generation.
type sortableTimestamps []int64
func (ts sortableTimestamps) Len() int           { return len(ts) }
func (ts sortableTimestamps) Less(i, j int) bool { return ts[i] < ts[j] }
func (ts sortableTimestamps) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

type commandInfo struct {
	// curCmd is used to record the digest of current command.
	curCmd string

	// priCmd is used to track the digests of the command which should be executed before current command.
	priCmd map[string]bool

	// pOrders is used to record the partial-order generated by phalanx replicas.
	pOrders []*protos.PartialOrder

	// timestamps is used to record the timestamp of partial orders.
	timestamps sortableTimestamps
}

func newCmdInfo(commandD string) *commandInfo {
	return &commandInfo{
		curCmd:  commandD,
		priCmd:  make(map[string]bool),
		pOrders: nil,
	}
}

func (info *commandInfo) pOrderAppend(pOrder *protos.PartialOrder) {
	info.pOrders = append(info.pOrders, pOrder)
	info.timestamps = append(info.timestamps, pOrder.Timestamp())
}

func (info *commandInfo) pOrderCount() int {
	return len(info.pOrders)
}

func (info *commandInfo) pOrderRead() []*protos.PartialOrder {
	return info.pOrders
}

func (info *commandInfo) prioriRecord(commandD string) {
	info.priCmd[commandD] = true
}

func (info *commandInfo) prioriCommit(commandD string) {
	delete(info.priCmd, commandD)
}

func (info *commandInfo) prioriFinished() bool {
	return len(info.priCmd) == 0
}