package finality

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
	"sort"
	"sync"
)

type timestampBasedOrdering struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.RWMutex

	//============================ basic information =============================================

	// author indicates the identifier of current node.
	author uint64

	// seqNo is used to generate sequential blocks.
	seqNo uint64

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// blocks is used to record the command info which could be committed.
	blocks types.SortableInnerBlocks

	//======================================= essential tools ===============================================

	// cRecorder is used to record the command info.
	cRecorder api.CommandRecorder

	// reader is used to read raw commands from meta pool.
	reader api.MetaReader

	// metrics is used to record the metric of timestamp-based ordering.
	metrics *metrics.OrderRuleMetrics

	// logger is used to print logs.
	logger external.Logger
}

func (tbo *timestampBasedOrdering) commitOrderStream(oStream types.OrderStream) {
	if len(oStream) == 0 {
		return
	}

	updated := false // if we have updated the command collector.
	for _, oInfo := range oStream {
		// order rule 1: collection rule, collect the partial order info.
		updated = tbo.collectPartials(oInfo)
	}

	if updated {
		// if the collector has been updated, try to process the committed partial orders.
		tbo.processPartialOrder()
	}
}

func (tbo *timestampBasedOrdering) collectPartials(oInfo types.OrderInfo) bool {
	// find the digest for current command the partial order refers to.
	commandD := oInfo.Command

	// check if current command has been committed or not.
	if tbo.cRecorder.IsCommitted(commandD) {
		tbo.logger.Debugf("[%d] committed command %s, ignore it", tbo.author, commandD)
		return false
	}

	// push back partial order into recorder.queue.
	if err := tbo.cRecorder.PushBack(oInfo); err != nil {
		tbo.logger.Errorf("[%d] push back partial order failed: %s", tbo.author, err)
		return false
	}

	// already committed by quorum replicas, then update the timestamp list.
	if tbo.cRecorder.IsQuorum(commandD) {
		// ignore the commands which have already reached quorum status.
		return false
	}

	// read command info from command cRecorder.
	info := tbo.cRecorder.ReadCommandInfo(commandD)
	info.OrderAppend(oInfo)

	// check the command status.
	switch info.OrderCount() {
	case tbo.oneCorrect:
		// current command has reached correct sequenced status.
		tbo.cRecorder.CorrectStatus(commandD)
		tbo.logger.Infof("[%d] found correct sequenced command %s", tbo.author, commandD)
	case tbo.quorum:
		// current command has reached quorum sequenced status.
		tbo.cRecorder.QuorumStatus(commandD)
		tbo.logger.Infof("[%d] found quorum sequenced command %s", tbo.author, commandD)
		info.UpdateTrustedTS(tbo.oneCorrect)
		rawCommand := tbo.reader.ReadCommand(info.Digest)
		block := types.NewInnerBlock(tbo.seqNo, false, rawCommand, info.TrustedTS)
		tbo.blocks = append(tbo.blocks, block)
	}
	return true
}

// processPartialOrder is used to process partial order with phalanx anchor-based ordering rules.
func (tbo *timestampBasedOrdering) processPartialOrder() {
	qInfos := tbo.cRecorder.ReadQSCInfos()

	if len(qInfos) < 100 {
		return
	}

	sort.Sort(qInfos)
}
