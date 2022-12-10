package types

import (
	"fmt"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/google/btree"
)

type OrderInfo struct {
	// Author indicates the generator of current order.
	Author uint64

	// Sequence indicates the sequential indicator.
	Sequence uint64

	// Command indicates the digest of command which current info is ordered for.
	Command string

	// Timestamp indicates the time when current order is generated.
	Timestamp int64

	//
	AfterQuorum bool
}

func OrderAttemptToOrderInfos(seqNo uint64, attempt *protos.OrderAttempt) ([]OrderInfo, uint64) {
	nodeID := attempt.NodeID

	commandList := attempt.CommandList()

	timestampList := attempt.TimestampList()

	var infos []OrderInfo

	for index, command := range commandList {
		timestamp := timestampList[index]
		seqNo++
		info := OrderInfo{Author: nodeID, Sequence: seqNo, Command: command, Timestamp: timestamp}
		infos = append(infos, info)
	}

	return infos, seqNo
}

func PartialOrderToOrderInfos(seqNo uint64, pOrder *protos.PartialOrder) ([]OrderInfo, uint64) {
	author := pOrder.Author()

	commandList := pOrder.CommandList()

	timestampList := pOrder.TimestampList()

	var infos []OrderInfo

	for index, command := range commandList {
		timestamp := timestampList[index]
		seqNo++
		info := OrderInfo{Author: author, Sequence: seqNo, Command: command, Timestamp: timestamp}
		infos = append(infos, info)
	}

	return infos, seqNo
}

func (info OrderInfo) Format() string {
	return fmt.Sprintf("[OrderInfo: author %d, sequence no. %d, command %s, timestamp %d]", info.Author, info.Sequence, info.Command, info.Timestamp)
}

func (info OrderInfo) Less(item btree.Item) bool {
	// for b-tree initiation
	return info.Sequence < (item.(OrderInfo)).Sequence
}

type OrderStream []OrderInfo

func (os OrderStream) Len() int      { return len(os) }
func (os OrderStream) Swap(i, j int) { os[i], os[j] = os[j], os[i] }
func (os OrderStream) Less(i, j int) bool {
	// query index for the same node, sort according to sequence number.
	if os[i].Sequence == os[j].Sequence {
		return os[i].Author < os[j].Author
	}

	// sort according to author.
	return os[i].Sequence < os[j].Sequence
}
