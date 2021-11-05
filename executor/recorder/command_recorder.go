package recorder

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

type commandRecorder struct {
	// author indicates the identifier of current node.
	author uint64

	// mapCmd is a map for tracking the command info, including the selected partial order, the priori commands,
	// the content of current command.
	mapCmd map[string]*types.CommandInfo

	// mapCSC is a map for CSC (correct sequenced command), which indicates there is at least one correct replica
	// has given a partial order for current command.
	mapCSC map[string]bool

	// mapQSC is a map for QSC (quorum sequenced command), which indicates there are legal amount replicas have given
	// a partial order for current command which could be used to decide the natural order among commands.
	mapQSC map[string]bool

	// mapCmt is a map for commands which have already been committed.
	mapCmt map[string]bool

	// mapWat is a map for commands which have already become QSC but have some priorities.
	mapWat map[string]bool

	// mapPri the potential priori relation recorder to update the mapWat at the same time the priori command committed.
	mapPri map[string][]*types.CommandInfo

	// leaves is a budget map to record the leaf nodes in current execution graph.
	// we could skip to scan the cyclic dependency for command info which is not a leaf node.
	leaves map[string]bool

	// logger is used to print logs.
	logger external.Logger
}

func NewCommandRecorder(author uint64, logger external.Logger) internal.CommandRecorder {
	return &commandRecorder{
		author: author,
		mapCmd: make(map[string]*types.CommandInfo),
		mapCSC: make(map[string]bool),
		mapQSC: make(map[string]bool),
		mapCmt: make(map[string]bool),
		mapWat: make(map[string]bool),
		mapPri: make(map[string][]*types.CommandInfo),
		leaves: make(map[string]bool),
		logger: logger,
	}
}

//=============================== read command info ============================================

func (recorder *commandRecorder) ReadCommandInfo(commandD string) *types.CommandInfo {
	info, ok := recorder.mapCmd[commandD]
	if !ok {
		info = types.NewCmdInfo(commandD)
		recorder.mapCmd[commandD] = info
	}
	return info
}

func (recorder *commandRecorder) ReadCSCInfos() []*types.CommandInfo {
	// select the correct sequenced commands.
	var commandInfos []*types.CommandInfo
	for digest := range recorder.mapCSC {
		commandInfos = append(commandInfos, recorder.ReadCommandInfo(digest))
	}
	return commandInfos
}

func (recorder *commandRecorder) ReadQSCInfos() []*types.CommandInfo {
	// when we try to read one quorum sequenced command from recorder, we should check the pri-command at first to make
	// sure there isn't any potential pri-command.
	//
	// here, the commands with potential priori are removed from QSC map temporarily, so that the commands in QSC map
	// always have a nil pri-command list.

	var commandInfos []*types.CommandInfo
	for digest := range recorder.mapQSC {
		if recorder.IsCommitted(digest) {
			continue
		}

		qCommandInfo := recorder.ReadCommandInfo(digest)
		commandInfos = append(commandInfos, qCommandInfo)
	}
	return commandInfos
}

func (recorder *commandRecorder) ReadWatInfos() []*types.CommandInfo {
	// select the commands which have already become QSC, but have some potential priority commands.
	var commandInfos []*types.CommandInfo
	for digest := range recorder.mapWat {
		commandInfos = append(commandInfos, recorder.ReadCommandInfo(digest))
	}
	return commandInfos
}

//=================================== update command status ========================================

func (recorder *commandRecorder) CorrectStatus(commandD string) {
	// append the command which has become CSC into mapCSC.
	// there is at least one correct partial order selected into pExecutor.
	recorder.mapCSC[commandD] = true
}

func (recorder *commandRecorder) QuorumStatus(commandD string) {
	// append the command which has become QSC into mapQSC.
	// there are quorum partial order selected into pExecutor.
	recorder.mapQSC[commandD] = true
	delete(recorder.mapCSC, commandD)
}

func (recorder *commandRecorder) CommittedStatus(commandD string) {
	recorder.mapCmt[commandD] = true
	delete(recorder.mapQSC, commandD)
	delete(recorder.mapCmd, commandD)

	recorder.prioriCommit(commandD)
	delete(recorder.mapPri, commandD)
}

func (recorder *commandRecorder) prioriCommit(commandD string) {
	// notify the post commands that its priority has been committed.
	for _, waitingInfo := range recorder.mapPri[commandD] {
		waitingInfo.PrioriCommit(commandD)
		recorder.logger.Debugf("[%d] %s committed potential pri-command %s", recorder.author, waitingInfo.Format(), commandD)

		if waitingInfo.PrioriFinished() {
			recorder.logger.Debugf("[%d] %s finished potential priori", recorder.author, waitingInfo.Format())
			recorder.mapQSC[waitingInfo.CurCmd] = true
			delete(recorder.mapWat, waitingInfo.CurCmd)
		}
	}
	delete(recorder.mapPri, commandD)
}

//==================================== get command status =============================================

func (recorder *commandRecorder) IsCommitted(commandD string) bool {
	return recorder.mapCmt[commandD]
}

func (recorder *commandRecorder) IsQuorum(commandD string) bool {
	return recorder.mapQSC[commandD] || recorder.mapWat[commandD]
}

//================================ management of leaf nodes =============================================

func (recorder *commandRecorder) AddLeaf(digest string) {
	recorder.leaves[digest] = true
}

func (recorder *commandRecorder) CutLeaf(info *types.CommandInfo) {
	delete(recorder.leaves, info.CurCmd)
}

func (recorder *commandRecorder) IsLeaf(digest string) bool {
	return recorder.leaves[digest]
}

//=========================== commands with potential byzantine order =================================

func (recorder *commandRecorder) PotentialByz(info *types.CommandInfo, newPriorities []string) {
	// remove the potential commands with potential byzantine order from QSC map.
	// put it into waiting map.
	delete(recorder.mapQSC, info.CurCmd)
	recorder.mapWat[info.CurCmd] = true

	// update the priority map for current QSC.
	for _, priori := range newPriorities {
		info.PriCmd[priori] = true

		recorder.mapPri[priori] = append(recorder.mapPri[priori], info)

		for digest, cmd := range recorder.mapCmd[priori].LowCmd {
			info.LowCmd[digest] = cmd
		}
	}
}
