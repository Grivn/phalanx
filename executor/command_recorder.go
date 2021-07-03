package executor

import (
	"fmt"
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

	// mapPri the potential priori relation recorder to update the mapWat at the same time the priori command committed.
	mapPri map[string][]string

	// logger is used to print logs.
	logger external.Logger
}

func newCommandRecorder(author uint64, logger external.Logger) *commandRecorder {
	return &commandRecorder{
		author: author,
		mapRaw: make(map[string]*protos.Command),
		mapCmd: make(map[string]*commandInfo),
		mapCSC: make(map[string]bool),
		mapQSC: make(map[string]bool),
		mapCmt: make(map[string]bool),
		mapPri: make(map[string][]string),
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
	var commandInfos []*commandInfo
	for digest := range recorder.mapCSC {
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

		// check if all the pri-commands have been committed.
		if !qCommandInfo.prioriFinished() {
			panic(fmt.Sprintf("[%d] unfinished priori command, %s", recorder.author, qCommandInfo.format()))
		}
	}
	return commandInfos
}

//=================================== update command status ========================================

func (recorder *commandRecorder) correctStatus(commandD string) {
	recorder.mapCSC[commandD] = true
}

func (recorder *commandRecorder) quorumStatus(commandD string) {
	recorder.mapQSC[commandD] = true
	delete(recorder.mapCSC, commandD)
}

func (recorder *commandRecorder) committedStatus(commandD string) {
	recorder.mapCmt[commandD] = true
	delete(recorder.mapQSC, commandD)
	delete(recorder.mapCmd, commandD)
	delete(recorder.mapRaw, commandD)

	recorder.prioriCommit(commandD)
}

//==================================== get command status =============================================

func (recorder *commandRecorder) isCommitted(commandD string) bool {
	return recorder.mapCmt[commandD]
}

//=========================== commands with potential byzantine order =================================

func (recorder *commandRecorder) potentialByz(info *commandInfo) {
	delete(recorder.mapQSC, info.curCmd)
	for priori  := range info.priCmd {
		recorder.mapPri[priori] = append(recorder.mapPri[priori], info.curCmd)
	}
}

func (recorder *commandRecorder) prioriCommit(commandD string) {
	afterList, ok := recorder.mapPri[commandD]
	if !ok {
		return
	}

	for _, digest := range afterList {
		waitingInfo := recorder.readCommandInfo(digest)
		waitingInfo.prioriCommit(commandD)
		recorder.logger.Debugf("[%d] %s committed potential pri-command %s", recorder.author, waitingInfo.format(), commandD)

		if waitingInfo.prioriFinished() {
			recorder.logger.Debugf("[%d] %s finished potential priori", recorder.author, waitingInfo.format())
			recorder.mapQSC[waitingInfo.curCmd] = true
		}
	}
}
