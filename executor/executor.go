package executor

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/internal"
	"github.com/google/btree"
	"sync"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type executorImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.Mutex

	// cacheTree
	cacheTree *btree.BTree

	// executedNo
	executedNo uint64

	//
	executionC chan *protos.PartialOrderBatch

	// author indicates the identifier of current node.
	author uint64

	// seqNo is used to track the sequence number for blocks.
	seqNo uint64

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	// recorder is used to record the command info.
	recorder *commandRecorder

	//
	closeC chan bool

	//
	mgr internal.LogManager

	// exec is used to execute the block.
	exec external.ExecutionService

	//
	logger external.Logger
}

// NewExecutor is used to generator an executor for phalanx.
func NewExecutor(author uint64, n int, mgr internal.LogManager, exec external.ExecutionService, logger external.Logger) *executorImpl {
	recorder := newCommandRecorder(author, logger)
	return &executorImpl{
		executedNo: uint64(0),
		cacheTree: btree.New(2),
		author:    author,
		rules:     newOrderRule(author, n, recorder, logger),
		recorder:  recorder,
		exec:      exec,
		mgr:       mgr,
		logger:    logger,
	}
}

func (ei *executorImpl) Run() {
	for {
		select {
		case <-ei.closeC:
			return
		case pBatch := <-ei.executionC:
			ei.logger.Infof("read %v", pBatch)
			ei.commitPartials(pBatch)
		default:
			continue
		}
	}
}

func (ei *executorImpl) Commit(event *types.CommitEvent) {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	ei.logger.Infof("[%d] phalanx received commit event %s", ei.author, event.Format())
	ei.cacheTree.ReplaceOrInsert(event)
	if ev := ei.minCommitEvent(); ev != nil {
		ei.commitPartials(ev.PBatch)
	}
}

func (ei *executorImpl) minCommitEvent() *types.CommitEvent {
	item := ei.cacheTree.Min()
	if item == nil {
		return nil
	}

	event, ok := item.(*types.CommitEvent)
	if !ok {
		return nil
	}

	if event.Sequence != ei.executedNo+1 {
		return nil
	}

	ei.executedNo++
	ei.cacheTree.Delete(event)

	if !event.IsValid {
		return ei.minCommitEvent()
	}

	return event
}

// CommitPartials is used to commit the QCs.
func (ei *executorImpl) commitPartials(pBatch *protos.PartialOrderBatch) {
	if pBatch == nil {
		// nil partial order batch means we should skip the current commitment attempt.
		return
	}

	for _, rawCommand := range pBatch.Commands {
		ei.recorder.storeCommand(rawCommand)
	}

	for _, pOrder := range pBatch.Partials {
		blocks := ei.rules.processPartialOrder(pOrder)
		for _, blk := range blocks {
			ei.seqNo++
			ei.exec.CommandExecution(blk.CommandD, blk.TxList, ei.seqNo, blk.Timestamp)
			ei.mgr.Committed(blk.Author, blk.CmdSeq)
		}
	}

	if ev := ei.minCommitEvent(); ev != nil {
		ei.commitPartials(ev.PBatch)
	}
}
