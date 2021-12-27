package internal

import (
	"github.com/Grivn/phalanx/common/types"
)

type Executor interface {
	// CommitStream is used to commit the partial order stream.
	CommitStream(qStream types.QueryStream) error

	// QueryMetrics returns metrics info of executor.
	QueryMetrics() types.MetricsInfo
}

//=============================================== Command Reader for Executor =====================================================

type CommandRecorder interface {
	InfoReader
	InfoStatus
	LeafManager
	PriorityManager
	QueueManager
}

type InfoReader interface {
	// ReadCommandInfo returns command with specific digest.
	ReadCommandInfo(commandD string) *types.CommandInfo

	// ReadCSCInfos returns commands in correct sequenced status.
	ReadCSCInfos() []*types.CommandInfo

	// ReadQSCInfos returns commands in quorum sequenced status.
	ReadQSCInfos() []*types.CommandInfo

	// ReadWatInfos returns commands which are waiting for priorities' commitment.
	ReadWatInfos() []*types.CommandInfo
}

type InfoStatus interface {
	// CorrectStatus set commands into correct sequenced status.
	CorrectStatus(commandD string)

	// QuorumStatus set commands into quorum sequenced status.
	QuorumStatus(commandD string)

	// CommittedStatus set commands into committed status.
	CommittedStatus(commandD string)

	// IsCommitted returns if current command has been committed.
	IsCommitted(commandD string) bool

	// IsQuorum returns if we have received quorum partial orders for this command.
	IsQuorum(commandD string) bool
}

type LeafManager interface {
	// AddLeaf adds a leaf command info node.
	AddLeaf(digest string)

	// CutLeaf cuts a leaf command info node.
	CutLeaf(info *types.CommandInfo)

	// IsLeaf returns if current command info is a leaf node.
	IsLeaf(digest string) bool
}

type PriorityManager interface {
	// PotentialByz add priorities for current command.
	PotentialByz(info *types.CommandInfo, newPriorities []string)
}

type QueueManager interface {
	// PushBack pushes the partial orders into FIFO order queue for each node.
	PushBack(oInfo types.OrderInfo) error

	// FrontCommands selects commands which is possible to be committed at first.
	FrontCommands() ([]string, bool)

	// OligarchyLeaderFront returns the oligarchy leader ordering, test mode.
	OligarchyLeaderFront(leader uint64) string
}

//================================== Cyclic Scanner ==============================================

type CondorcetScanner interface {
	HasCyclic() bool
}
