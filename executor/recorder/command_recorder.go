package recorder

import (
	"container/list"
	"fmt"
	"github.com/Grivn/phalanx/common/protos"
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

	// FIFOQueue is used to record partial orders from each node.
	FIFOQueue map[uint64]*list.List

	//
	oneCorrect int

	quorum int

	// logger is used to print logs.
	logger external.Logger
}

func NewCommandRecorder(author uint64, n int, logger external.Logger) internal.CommandRecorder {
	set := make(map[uint64]*list.List)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		set[id] = list.New()
	}
	return &commandRecorder{
		author: author,
		mapCmd: make(map[string]*types.CommandInfo),
		mapCSC: make(map[string]bool),
		mapQSC: make(map[string]bool),
		mapCmt: make(map[string]bool),
		mapWat: make(map[string]bool),
		mapPri: make(map[string][]*types.CommandInfo),
		leaves: make(map[string]bool),
		oneCorrect: types.CalculateOneCorrect(n),
		quorum: types.CalculateQuorum(n),
		FIFOQueue: set,
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

func (recorder *commandRecorder) IsCorrect(commandD string) bool {
	return recorder.mapCSC[commandD]
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

//=================================== queue manager ===============================================

// PushBack pushes the partial orders into FIFO order queue for each node.
func (recorder *commandRecorder) PushBack(pOrder *protos.PartialOrder) error {
	if recorder.IsCommitted(pOrder.CommandDigest()) {
		// ignore committed command.
		return nil
	}

	queue, ok := recorder.FIFOQueue[pOrder.Author()]

	if !ok {
		return fmt.Errorf("cannot find order queue of node %d", pOrder.Author())
	}

	queue.PushBack(pOrder)
	return nil
}

// FrontCommands selects commands which is possible to be committed at first.
func (recorder *commandRecorder) FrontCommands() []string {
	var fronts []string

	counts := make(map[string]int)

	for _, queue := range recorder.FIFOQueue {

		for {
			if queue.Len() == 0 {
				break
			}

			e := queue.Front()

			pOrder, ok := e.Value.(*protos.PartialOrder)

			if !ok {
				queue.Remove(e)
				continue
			}

			if recorder.IsCommitted(pOrder.CommandDigest()) {
				queue.Remove(e)
				continue
			}

			fronts = append(fronts, pOrder.CommandDigest())

			counts[pOrder.CommandDigest()]++

			recorder.logger.Infof("[%d] select node %d front command %s", recorder.author, pOrder.Author(), pOrder.CommandDigest())
			break
		}
	}

	if len(fronts) < recorder.quorum {
		// less than quorum participants provide partial order queue,
		// we cannot find any command info in quorum status, just return nil list.
		return nil
	}

	var correct []string

	for digest, count := range counts {
		if count < recorder.oneCorrect {
			continue
		}
		correct = append(correct, digest)
	}

	if len(correct) == 0 {
		// we cannot find any command in correct status, just return all the front command digests.
		var unverified []string
		for digest := range counts {
			unverified = append(unverified, digest)
		}
		recorder.logger.Debugf("[%d] unverified front digest %v", recorder.author, unverified)
		return recorder.frontFilter(unverified)
	}

	recorder.logger.Debugf("[%d] correct front digest %v", recorder.author, correct)
	return recorder.frontFilter(correct)
}

func (recorder *commandRecorder) frontFilter(fronts []string) []string {
	if len(fronts) == 0 {
		return nil
	}

	if len(fronts) == 1 {
		// don't need to filter front digest.
		return fronts
	}

	var filtered []string

	for _, digest := range fronts {
		if recorder.IsCorrect(digest) {
			// we should make sure that there aren't any commands in correct status.
			recorder.logger.Debugf("[%d] found correct status command %s, skip execution", recorder.author, digest)
			filtered = nil
			break
		}

		if recorder.IsQuorum(digest) {
			// we should select the command ordered by at least quorum participants.
			recorder.logger.Debugf("[%d] select quorum status command %s", recorder.author, digest)
			filtered = append(filtered, digest)
		}
	}

	recorder.logger.Infof("[%d] filtered front digest %v", recorder.author, filtered)
	return filtered
}
