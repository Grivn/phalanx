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
	mapPri map[string]map[string]*types.CommandInfo

	// leaves is a budget map to record the leaf nodes in current execution graph.
	// we could skip to scan the cyclic dependency for command info which is not a leaf node.
	leaves map[string]bool

	leafGroup map[string]map[string]*types.CommandInfo

	traceLeaf map[string]map[string]bool

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
		mapPri: make(map[string]map[string]*types.CommandInfo),
		leafGroup: make(map[string]map[string]*types.CommandInfo),
		leaves: make(map[string]bool),
		traceLeaf: make(map[string]map[string]bool),
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
	recorder.leafGroup[digest] = make(map[string]*types.CommandInfo)
	recorder.traceLeaf[digest] = make(map[string]bool)
}

func (recorder *commandRecorder) CutLeaf(info *types.CommandInfo) {
	delete(recorder.leaves, info.CurCmd)
	delete(recorder.leafGroup, info.CurCmd)
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
		// record the priority command.
		info.PriCmd[priori] = true

		recorder.updatePrioriMap(priori, info)

		// the priority command is a leaf node.
		if recorder.IsLeaf(priori) {
			lGroup := recorder.leafGroup[priori]
			lGroup[info.CurCmd] = info
			recorder.traceLeaf[info.CurCmd][priori] = true

			// current command is a leaf node.
			if recorder.IsLeaf(info.CurCmd) {
				curLGroup := recorder.leafGroup[info.CurCmd]
				for _, v := range curLGroup {
					lGroup[v.CurCmd] = v
					recorder.traceLeaf[v.CurCmd][priori] = true
				}
			}
		} else {
			for leafD := range recorder.traceLeaf[priori] {
				if !recorder.IsLeaf(leafD) {
					continue
				}
				lGroup := recorder.leafGroup[leafD]
				recorder.logger.Debugf("trace leaf group %s to %s", priori, leafD)
				lGroup[info.CurCmd] = info
				recorder.traceLeaf[info.CurCmd][leafD] = true
			}
		}

		// update current node low command map.
		//info.TransitiveLow(recorder.mapCmd[priori])
		//for digest, cmd := range recorder.mapCmd[priori].LowCmd {
		//	// record the priority execution tracing map.
		//	//recorder.mapPri[digest] = append(recorder.mapPri[digest], info)
		//	recorder.updatePrioriMap(digest, info)
		//
		//	info.LowCmd[digest] = cmd
		//}
	}

	//
	//// update the low command map of current command's children.
	//if recorder.leaves[info.CurCmd] {
	//	lGroup := recorder.leafGroup[info.CurCmd]
	//
	//	if lGroup == nil {
	//		return
	//	}
	//
	//	for _, nextCmd := range lGroup {
	//		// remove current QSC from its children's low command map.
	//		recorder.logger.Debugf("[%d] update children's low map, %s", recorder.author, nextCmd.Format())
	//
	//		// append the low commands of current QSC info into its children's low map.
	//		for digest, cmd := range info.LowCmd {
	//			// record the priority execution tracing map.
	//			//recorder.mapPri[digest] = append(recorder.mapPri[digest], info)
	//			recorder.updatePrioriMap(digest, info)
	//
	//			nextCmd.LowCmd[digest] = cmd
	//		}
	//	}
	//}
}

func (recorder *commandRecorder) CheckLeaves(curInfo, checkInfo *types.CommandInfo) bool {
	lGroup := recorder.leafGroup[curInfo.CurCmd]

	_, ok := lGroup[checkInfo.CurCmd]
	return ok
}

func (recorder *commandRecorder) updatePrioriMap(priori string, info *types.CommandInfo) {
	// record the priority execution tracing map.
	prioriMap, ok := recorder.mapPri[priori]
	if !ok {
		prioriMap = make(map[string]*types.CommandInfo)
		recorder.mapPri[priori] = prioriMap
	}
	prioriMap[info.CurCmd] = info
}
