package types

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// sortableTimestamps is a sortable slice for free will trusted timestamp generation.
type sortableTimestamps []int64

func (ts sortableTimestamps) Len() int           { return len(ts) }
func (ts sortableTimestamps) Less(i, j int) bool { return ts[i] < ts[j] }
func (ts sortableTimestamps) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

type CommandInfo struct {
	// Digest is used to record the digest of current command.
	Digest string

	// PriCmd is used to track the digests of the command which should be executed before current command.
	PriCmd map[string]bool

	// LowCmd is used to record the lowest commands which could be regarded as the first priority command for us.
	LowCmd map[string]*CommandInfo

	// Orders is used to record the partial-order generated by phalanx replicas.
	Orders map[uint64]OrderInfo

	// Timestamps is used to record the timestamp of partial orders.
	Timestamps sortableTimestamps

	// Trust indicates if we have checked the potential priori command for it.
	Trust bool

	// GTime is the timestamp to generate current command info.
	GTime int64

	// TrustedTS is the timestamp current command came into system.
	TrustedTS int64
}

func (ci *CommandInfo) UpdateTrustedTS(oneCorrect int) {
	sort.Sort(ci.Timestamps)
	ci.TrustedTS = ci.Timestamps[oneCorrect-1]
}

type CommandStream []*CommandInfo

func NewCmdInfo(commandD string) *CommandInfo {
	return &CommandInfo{
		Digest: commandD,
		PriCmd: make(map[string]bool),
		LowCmd: make(map[string]*CommandInfo),
		Orders: make(map[uint64]OrderInfo),
		Trust:  false,
		GTime:  time.Now().UnixNano(),
	}
}

func (s CommandStream) Len() int      { return len(s) }
func (s CommandStream) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s CommandStream) Less(i, j int) bool {
	if s[i].TrustedTS == s[j].TrustedTS {
		return s[i].Digest < s[j].Digest
	}
	return s[i].TrustedTS < s[j].TrustedTS
}

func (ci *CommandInfo) Format() string {
	set := make([]string, 0, len(ci.Orders))
	for _, pOrder := range ci.Orders {
		set = append(set, fmt.Sprintf("<%d, %d>", pOrder.Author, pOrder.Timestamp))
	}
	return fmt.Sprintf("[CommandInfo: command %s, order-count %d, trusted-ts %d, <tss %s>]", ci.Digest, len(ci.Orders), ci.TrustedTS, strings.Join(set, ","))
}

//========================== Partial Order Manager ====================================

func (ci *CommandInfo) OrderAppend(oInfo OrderInfo) {
	ci.Orders[oInfo.Author] = oInfo
	ci.Timestamps = append(ci.Timestamps, oInfo.Timestamp)
}

func (ci *CommandInfo) OrderCount() int {
	return len(ci.Orders)
}

//========================== Priority Command ====================================

func (ci *CommandInfo) PrioriRecord(priInfo *CommandInfo) {
	ci.PriCmd[priInfo.Digest] = true
}

func (ci *CommandInfo) PrioriCommit(commandD string) {
	delete(ci.PriCmd, commandD)
}

func (ci *CommandInfo) PrioriFinished() bool {
	return len(ci.PriCmd) == 0
}

//=========================== Lowest Command ====================================

// TransitiveLow update current node's lowest map.
// <x,y> && <y,z> -> <x,z>
func (ci *CommandInfo) TransitiveLow(parentInfo *CommandInfo) {
	// remove the parent partial order from lowest map.
	delete(ci.LowCmd, parentInfo.Digest)

	// append parent's lowest into ourselves.
	for _, info := range parentInfo.LowCmd {
		ci.AppendLow(info)
	}
}

func (ci *CommandInfo) AppendLow(info *CommandInfo) {
	// append partial order into our lowest list.
	ci.LowCmd[info.Digest] = info
}
