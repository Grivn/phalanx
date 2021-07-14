package executor

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type commandRecorder struct {
	// author indicates the identifier of current node.
	author uint64

	// mapRaw is a map for tracking the command content.
	mapRaw map[string]*protos.Command

	// mapCmd is a map for tracking the command info, including the selected partial order, the priori commands,
	// the content of current command.
	mapCmd map[string]*commandInfo

	// mapCSC is a map for CSC (correct sequenced command), which indicates there is at least one correct replica
	// has given a partial order for current command.
	mapCSC map[string]bool

	// mapQSC is a map for QSC (quorum sequenced command), which indicates there are legal amount replicas have given
	// a partial order for current command which could be used to decide the natural order among commands.
	mapQSC map[string]bool

	// mapCmt is a map for commands which have already been committed.
	mapCmt map[string]bool

	// todo reduce the check-time for the command pairs with potential natural order, mapPri & mapWat.

	mapWat map[string]bool

	// mapPri the potential priori relation recorder to update the mapWat at the same time the priori command committed.
	mapPri map[string]map[string]bool

	// todo update the leaf node algorithm to find cyclic dependency.
	mapLeaf map[string]bool

	// logger is used to print logs.
	logger external.Logger
}

type forestGroup struct {
	components map[string]bool
	leaves     map[string]bool
}

func newCommandRecorder(author uint64, logger external.Logger) *commandRecorder {
	return &commandRecorder{
		author: author,
		mapRaw: make(map[string]*protos.Command),
		mapCmd: make(map[string]*commandInfo),
		mapCSC: make(map[string]bool),
		mapQSC: make(map[string]bool),
		mapCmt: make(map[string]bool),
		mapWat: make(map[string]bool),
		mapPri: make(map[string]map[string]bool),
		mapLeaf: make(map[string]bool),
		logger: logger,
	}
}

//=============================== store raw data ===============================================

func (recorder *commandRecorder) storeCommand(command *protos.Command) {
	if recorder.mapCmt[command.Digest] {
		return
	}
	recorder.mapRaw[command.Digest] = command
}

//=============================== read command info ============================================

func (recorder *commandRecorder) readCommandRaw(commandD string) *protos.Command {
	return recorder.mapRaw[commandD]
}

func (recorder *commandRecorder) readCommandInfo(commandD string) *commandInfo {
	info, ok := recorder.mapCmd[commandD]
	if !ok {
		info = newCmdInfo(commandD)
		recorder.mapCmd[commandD] = info
	}
	return info
}

func (recorder *commandRecorder) readCSCInfos() []*commandInfo {
	// select the correct sequenced commands.
	var commandInfos []*commandInfo
	for digest := range recorder.mapCSC {
		commandInfos = append(commandInfos, recorder.readCommandInfo(digest))
	}
	return commandInfos
}

func (recorder *commandRecorder) readWatInfos() []*commandInfo {
	// select the commands which have already become QSC, but have some potential priority commands.
	var commandInfos []*commandInfo
	for digest := range recorder.mapWat {
		commandInfos = append(commandInfos, recorder.readCommandInfo(digest))
	}
	return commandInfos
}

func (recorder *commandRecorder) readQSCInfos() []*commandInfo {
	// when we try to read one quorum sequenced command from recorder, we should check the pri-command at first to make
	// sure there isn't any potential pri-command.
	//
	// here, the commands with potential priori are removed from QSC map temporarily, so that the commands in QSC map
	// always have a nil pri-command list.

	var commandInfos []*commandInfo
	for digest := range recorder.mapQSC {
		if recorder.isCommitted(digest) {
			continue
		}

		qCommandInfo := recorder.readCommandInfo(digest)
		commandInfos = append(commandInfos, qCommandInfo)
	}
	return commandInfos
}

//=================================== update command status ========================================

func (recorder *commandRecorder) correctStatus(commandD string) {
	// append the command which has become CSC into mapCSC.
	// there is at least one correct partial order selected into pExecutor.
	recorder.mapCSC[commandD] = true
}

func (recorder *commandRecorder) quorumStatus(commandD string) {
	// append the command which has become QSC into mapQSC.
	// there are quorum partial order selected into pExecutor.
	recorder.mapQSC[commandD] = true
	delete(recorder.mapCSC, commandD)
}

func (recorder *commandRecorder) committedStatus(commandD string) {
	recorder.mapCmt[commandD] = true
	delete(recorder.mapQSC, commandD)
	delete(recorder.mapCmd, commandD)
	delete(recorder.mapRaw, commandD)

	recorder.prioriCommit(commandD)
	delete(recorder.mapPri, commandD)
}

//==================================== get command status =============================================

func (recorder *commandRecorder) isCommitted(commandD string) bool {
	return recorder.mapCmt[commandD]
}

//=========================== commands with potential byzantine order =================================

func (recorder *commandRecorder) potentialByz(info *commandInfo) {
	// remove the potential commands with potential byzantine order from QSC map.
	// put it into waiting map.
	delete(recorder.mapQSC, info.curCmd)
	recorder.mapWat[info.curCmd] = true

	// update the priority map for current QSC.
	for priori := range info.priCmd {
		m, ok := recorder.mapPri[priori]
		if !ok {
			m = make(map[string]bool)
			recorder.mapPri[priori] = m
		}

		m[info.curCmd] = true
	}
}

func (recorder *commandRecorder) prioriCommit(commandD string) {
	m, ok := recorder.mapPri[commandD]
	if !ok {
		return
	}

	for digest := range m {
		waitingInfo := recorder.readCommandInfo(digest)
		waitingInfo.prioriCommit(commandD)
		recorder.logger.Debugf("[%d] %s committed potential pri-command %s", recorder.author, waitingInfo.format(), commandD)

		if waitingInfo.prioriFinished() {
			recorder.logger.Debugf("[%d] %s finished potential priori", recorder.author, waitingInfo.format())
			recorder.mapQSC[waitingInfo.curCmd] = true
			delete(recorder.mapWat, waitingInfo.curCmd)
		}
	}
}

//===============================================================

type scanner struct {
	recorder *commandRecorder

	priority *commandInfo

	sDigest string

	sDependent bool
}

func newScanner(recorder *commandRecorder, priInfo *commandInfo, selfCmd string) *scanner {
	return &scanner{recorder: recorder, priority: priInfo, sDigest: selfCmd, sDependent: false}
}

func (scanner *scanner) scan() bool {
	scanner.selfDependency(scanner.priority)
	return scanner.sDependent
}

func (scanner *scanner) selfDependency(priInfo *commandInfo) {
	if scanner.sDependent == true {
		return
	}

	if !scanner.recorder.mapWat[priInfo.curCmd] {
		if priInfo.curCmd == scanner.sDigest {
			scanner.sDependent = true
		}
		return
	}

	for digest := range priInfo.priCmd {
		if scanner.sDependent == true {
			return
		}
		scanner.selfDependency(scanner.recorder.readCommandInfo(digest))
	}
}

//===================================================================

//func (recorder *commandRecorder) isLeaf(commandD string) bool {
//	return recorder.mapLeaf[commandD]
//}
//
//func (recorder *commandRecorder) recordLeaf(commandD string) {
//	recorder.mapLeaf[commandD] = true
//}
//
//func (recorder *commandRecorder) removeLeaf(commandD string) {
//	delete(recorder.mapLeaf, commandD)
//}
//
//func (recorder *commandRecorder) getForest(digest string) *forestGroup {
//	forest, ok := recorder.mapNode[digest]
//	if !ok {
//		forest = &forestGroup{components: make(map[string]bool), leaves: make(map[string]bool)}
//		recorder.mapNode[digest] = forest
//	}
//	return forest
//}
