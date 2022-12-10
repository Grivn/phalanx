package finality

import (
	"sort"

	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/utils/interceptor"
	"github.com/Grivn/phalanx/pkg/utils/recorder"
	"github.com/google/btree"
)

type timestampAnchorBasedOrdering struct {
	//============================== basic info =====================================

	// author indicates the identifier of current node.
	author uint64

	// seqNo indicates the order of inner blocks.
	seqNo uint64

	// fault indicates the max amount byzantine node in current cluster.
	fault int

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// oligarchy is used to define that current cluster is relying on a certain node.
	oligarchy uint64

	// frontNo is used to track the sequence number for front stream.
	frontNo uint64

	//============================= internal interfaces =========================================

	// reload is used to notify client instance the committed sequence number.
	reload api.MetaCommitter

	// cRecorder is used to record the command info.
	cRecorder api.CommandRecorder

	// democracy is used to generate block with free will committee.
	democracy map[uint64]*btree.BTree

	// reader is used to read raw commands from meta pool.
	reader api.MetaReader

	//============================== external interfaces ==========================================

	// exec is used to execute the block.
	exec external.Executor

	// logger is used to print logs.
	logger external.Logger

	// metrics is used to record the metric info of current node's order rule module.
	metrics *metrics.ManipulationMetrics
}

func newTimestampAnchorBasedOrdering(
	conf config.PhalanxConf,
	meta api.MetaPool,
	executor external.Executor,
	logger external.Logger,
	ms *metrics.Metrics) *timestampAnchorBasedOrdering {
	democracy := make(map[uint64]*btree.BTree)
	for i := 0; i < conf.NodeCount; i++ {
		democracy[uint64(i+1)] = btree.New(2)
	}
	return &timestampAnchorBasedOrdering{
		author:     conf.NodeID,
		fault:      types.CalculateFault(conf.NodeCount),
		oneCorrect: types.CalculateOneCorrect(conf.NodeCount),
		quorum:     types.CalculateQuorum(conf.NodeCount),
		oligarchy:  conf.OligarchID,
		frontNo:    uint64(0),
		reload:     meta,
		cRecorder:  recorder.NewCommandRecorder(conf.NodeID, conf.NodeCount, logger),
		reader:     meta,
		democracy:  democracy,
		exec:       executor,
		logger:     logger,
		metrics:    ms.TimestampAnchorMetrics,
	}
}

func (tab *timestampAnchorBasedOrdering) commitOrderStream(oStream types.OrderStream) {
	if len(oStream) == 0 {
		return
	}

	updated := false // if we have updated the command collector.
	for _, oInfo := range oStream {
		// order rule 1: collection rule, collect the partial order info.
		updated = tab.collectPartials(oInfo)
	}

	if updated {
		// if the collector has been updated, try to process the committed partial orders.
		tab.processPartialOrder()
	}
}

// processPartialOrder is used to process partial order with phalanx anchor-based ordering rules.
func (tab *timestampAnchorBasedOrdering) processPartialOrder() {
	for {
		// order rule 2: execution rule, select commands to execute with natural order.
		anchorSet := tab.fetchAnchorSet()

		// order rule 3: commitment rule, generate ordered blocks with free will.
		blocks, frontNo := tab.freeWill(anchorSet)
		if len(blocks) == 0 {
			// there isn't a committed inner block.
			break
		}

		// commit blocks.
		tab.logger.Debugf("[%d] commit front group, front-no. %d, safe %v, blocks count %d", tab.author, frontNo, anchorSet.Safe, len(blocks))
		for _, blk := range blocks {
			tab.seqNo++
			tab.exec.CommandExecution(blk, tab.seqNo)
			tab.reload.Committed(blk.Command.Author, blk.Command.Sequence)

			// record metrics.
			tab.metrics.CommitBlock(blk)
		}
	}
}

func (tab *timestampAnchorBasedOrdering) collectPartials(oInfo types.OrderInfo) bool {
	// collect indicates the collection rule of phalanx:
	// which partial orders would be selected into execution process to compare order.
	tab.logger.Infof("[%d] collect partial order: %s", tab.author, oInfo.Format())

	// find the digest for current command the partial order refers to.
	commandD := oInfo.Command

	// check if current command has been committed or not.
	if tab.cRecorder.IsCommitted(commandD) {
		tab.logger.Debugf("[%d] committed command %s, ignore it", tab.author, commandD)
		return false
	}

	// push back partial order into recorder.queue.
	if err := tab.cRecorder.PushBack(oInfo); err != nil {
		tab.logger.Errorf("[%d] push back partial order failed: %s", tab.author, err)
		return false
	}

	// read command info from command cRecorder.
	info := tab.cRecorder.ReadCommandInfo(commandD)
	info.OrderAppend(oInfo)

	// already committed by quorum replicas, then update the timestamp list.
	if tab.cRecorder.IsQuorum(commandD) {
		info.UpdateTrustedTS(tab.oneCorrect)
	}

	// check the command status.
	switch info.OrderCount() {
	case tab.oneCorrect:
		// current command has reached correct sequenced status.
		tab.cRecorder.CorrectStatus(commandD)
		tab.logger.Infof("[%d] found correct sequenced command %s", tab.author, commandD)
	case tab.quorum:
		// current command has reached quorum sequenced status.
		tab.cRecorder.QuorumStatus(commandD)
		tab.logger.Infof("[%d] found quorum sequenced command %s", tab.author, commandD)
		info.UpdateTrustedTS(tab.oneCorrect)
	}
	return true
}

func (tab *timestampAnchorBasedOrdering) fetchAnchorSet() types.FrontStream {
	// execute indicates the execution rule of phalanx:
	// which commands would be selected into commitment process to generate blocks.
	// here, we should take 'Natural Order' into thought.

	// oligarchy mode, relying on certain leader ordering.
	if tab.oligarchy != uint64(0) {
		return tab.oligarchyExecution()
	}

	// read the front set.
	var cStream types.CommandStream
	if qInfo := tab.cRecorder.PickQuorumInfo(); qInfo != nil {
		// we cannot make sure the validation of front set.
		cStream = interceptor.NewInterceptor(tab.author, tab.cRecorder, tab.oneCorrect, tab.logger).SelectToCommit(types.CommandStream{qInfo})
	}

	return types.FrontStream{Safe: false, Stream: cStream}
}

func (tab *timestampAnchorBasedOrdering) oligarchyExecution() types.FrontStream {
	digest := tab.cRecorder.OligarchyLeaderFront(tab.oligarchy)
	commandInfo := tab.cRecorder.ReadCommandInfo(digest)
	if len(commandInfo.Orders) < tab.quorum {
		return types.FrontStream{Safe: true, Stream: nil}
	}
	return types.FrontStream{Safe: true, Stream: types.CommandStream{commandInfo}}
}

func (tab *timestampAnchorBasedOrdering) freeWill(frontStream types.FrontStream) ([]types.InnerBlock, uint64) {
	// commit indicates the commitment rule of phalanx:
	// generate blocks and assign sequence order for them.
	// here, the block generation would follow the 'Free Will' of participants.
	if len(frontStream.Stream) == 0 {
		return nil, tab.frontNo
	}

	tab.frontNo++

	// free will:
	// generate blocks and sort according to the trusted timestamp
	// here, the command-pair with natural order cannot take part in concurrent command set.
	var sortable types.SortableInnerBlocks
	for _, frontC := range frontStream.Stream {
		// record metrics.
		//tab.rMetrics.CommitFrontCommandInfo(frontC)

		// generate block, try to fetch the raw command to fulfill the block.
		rawCommand := tab.reader.ReadCommand(frontC.Digest)
		block := types.NewInnerBlock(tab.frontNo, frontStream.Safe, rawCommand, frontC.TrustedTS)
		tab.logger.Infof("[%d] generate block %s", tab.author, block.Format())

		// finished the block generation for command (digest), update the status of digest in command recorder.
		tab.cRecorder.CommittedStatus(frontC.Digest)

		// append the current block into sortable slice, waiting for order-determination.
		sortable = append(sortable, block)
	}

	// determine the order of commands which do not have any natural orders according to trusted timestamp.
	sort.Sort(sortable)

	return sortable, tab.frontNo
}
