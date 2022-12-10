package finality

import (
	"sort"

	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/utils/recorder"
)

type timestampBasedOrdering struct {
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

	// reload is used to notify client instance the committed sequence number.
	reload api.MetaCommitter

	// reader is used to read raw commands from meta pool.
	reader api.MetaReader

	// metrics is used to record the metric of timestamp-based ordering.
	metrics *metrics.ManipulationMetrics

	// logger is used to print logs.
	logger external.Logger
}

func newTimestampBasedOrdering(
	conf config.PhalanxConf,
	meta api.MetaPool,
	logger external.Logger,
	ms *metrics.Metrics) *timestampBasedOrdering {
	return &timestampBasedOrdering{
		author:     conf.NodeID,
		seqNo:      uint64(1),
		oneCorrect: types.CalculateOneCorrect(conf.NodeCount),
		quorum:     types.CalculateQuorum(conf.NodeCount),
		cRecorder:  recorder.NewCommandRecorder(conf.NodeID, conf.NodeCount, logger),
		reader:     meta,
		reload:     meta,
		metrics:    ms.TimestampBasedMetrics,
		logger:     logger,
	}
}

func (tb *timestampBasedOrdering) commitOrderStream(oStream types.OrderStream) {
	if len(oStream) == 0 {
		return
	}

	updated := false // if we have updated the command collector.
	for _, oInfo := range oStream {
		// order rule 1: collection rule, collect the partial order info.
		updated = tb.collectPartials(oInfo)
	}

	if updated {
		// if the collector has been updated, try to process the committed partial orders.
		tb.processPartialOrder()
	}
}

func (tb *timestampBasedOrdering) collectPartials(oInfo types.OrderInfo) bool {
	// find the digest for current command the partial order refers to.
	commandD := oInfo.Command

	// check if current command has been committed or not.
	if tb.cRecorder.IsCommitted(commandD) {
		tb.logger.Debugf("[%d] committed command %s, ignore it", tb.author, commandD)
		return false
	}

	// push back partial order into recorder.queue.
	if err := tb.cRecorder.PushBack(oInfo); err != nil {
		tb.logger.Errorf("[%d] push back partial order failed: %s", tb.author, err)
		return false
	}

	// already committed by quorum replicas, then update the timestamp list.
	if tb.cRecorder.IsQuorum(commandD) {
		// ignore the commands which have already reached quorum status.
		return false
	}

	// read command info from command cRecorder.
	info := tb.cRecorder.ReadCommandInfo(commandD)
	info.OrderAppend(oInfo)

	// check the command status.
	switch info.OrderCount() {
	case tb.oneCorrect:
		// current command has reached correct sequenced status.
		tb.cRecorder.CorrectStatus(commandD)
		tb.logger.Infof("[%d] found correct sequenced command %s", tb.author, commandD)
	case tb.quorum:
		// current command has reached quorum sequenced status.
		tb.cRecorder.QuorumStatus(commandD)
		tb.logger.Infof("[%d] found quorum sequenced command %s", tb.author, commandD)
		info.UpdateTrustedTS(tb.oneCorrect)
		rawCommand := tb.reader.ReadCommand(info.Digest)
		block := types.NewInnerBlock(tb.seqNo, false, rawCommand, info.TrustedTS)
		tb.blocks = append(tb.blocks, block)
	}
	return true
}

// processPartialOrder is used to process partial order with phalanx anchor-based ordering rules.
func (tb *timestampBasedOrdering) processPartialOrder() {
	qInfos := tb.cRecorder.ReadQSCInfos()

	if len(qInfos) < 2 {
		return
	}

	sort.Sort(tb.blocks)
	commitBlocks := types.SortableInnerBlocks{}
	updateBlocks := types.SortableInnerBlocks{}
	for index, blk := range tb.blocks {
		if index < 2 {
			commitBlocks = append(commitBlocks, blk)
			continue
		}
		updateBlocks = append(updateBlocks, blk)
	}
	tb.blocks = updateBlocks

	for _, blk := range commitBlocks {
		tb.metrics.CommitBlock(blk)
		tb.reload.Committed(blk.Command.Author, blk.Command.Sequence)
	}
}
