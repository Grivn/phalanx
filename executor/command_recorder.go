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

	// mapPri the potential priori relation recorder to update the mapWat at the same time the priori command committed.
	mapPri map[string]map[string]bool

	// mapWat the commands at the first time reach quorum sequenced status found a potential priori command.
	mapWat map[string]*commandInfo

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
	// when we try to read one quorum sequenced command from recorder, we should check the pri-command at first:
	// 1) remove the committed commands from their pri-command waiting list;
	// 2) append the command to the returned value, if we do not need to wait for any pri-command commitment.

	var commandInfos []*commandInfo
	for digest := range recorder.mapQSC {
		if recorder.isCommitted(digest) {
			continue
		}

		qCommandInfo := recorder.readCommandInfo(digest)
		recorder.logger.Infof("[%d] check quorum sequenced info %s", recorder.author, qCommandInfo.format())

		// check if one pri-command has been committed or not.
		for priDigest := range qCommandInfo.priCmd {
			if recorder.isCommitted(priDigest) {
				recorder.logger.Debugf("[%d]    committed priori command %d", recorder.author, priDigest)
				qCommandInfo.prioriCommit(priDigest)
			}
		}

		// check if all the pri-commands have been committed.
		if qCommandInfo.prioriFinished() {
			recorder.logger.Debugf("[%d]    finished priori command", recorder.author)
			commandInfos = append(commandInfos, recorder.readCommandInfo(digest))
		} else {
			recorder.logger.Debugf("[%d]    unfinished priori command", recorder.author)
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
}

//==================================== get command status =============================================

func (recorder *commandRecorder) isCommitted(commandD string) bool {
	return recorder.mapCmt[commandD]
}
