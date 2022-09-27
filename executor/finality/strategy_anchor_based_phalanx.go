package finality

import (
	"sort"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/interceptor"
	"github.com/Grivn/phalanx/executor/recorder"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
	"github.com/google/btree"
)

type phalanxAnchorBasedOrdering struct {
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
	exec external.ExecutionService

	// logger is used to print logs.
	logger external.Logger

	// metrics is used to record the metric info of current node's order rule module.
	metrics *metrics.ManipulationMetrics

	// cMetrics is used to record the metric to commit command info in phalanx anchor-based ordering.
	cMetrics *metrics.CommitmentMetrics
}

func newPhalanxAnchorBasedOrdering(conf Config) *phalanxAnchorBasedOrdering {
	democracy := make(map[uint64]*btree.BTree)
	for i := 0; i < conf.N; i++ {
		democracy[uint64(i+1)] = btree.New(2)
	}
	return &phalanxAnchorBasedOrdering{
		author:     conf.Author,
		fault:      types.CalculateFault(conf.N),
		oneCorrect: types.CalculateOneCorrect(conf.N),
		quorum:     types.CalculateQuorum(conf.N),
		oligarchy:  conf.OLeader,
		frontNo:    uint64(0),
		reload:     conf.Pool,
		cRecorder:  recorder.NewCommandRecorder(conf.Author, conf.N, conf.Logger),
		reader:     conf.Pool,
		democracy:  democracy,
		exec:       conf.Exec,
		logger:     conf.Logger,
		metrics:    conf.Metrics.PhalanxAnchorMetrics,
		cMetrics:   conf.Metrics.CommitmentMetrics,
	}
}

func (pab *phalanxAnchorBasedOrdering) commitOrderStream(oStream types.OrderStream) {
	if len(oStream) == 0 {
		return
	}

	updated := false // if we have updated the command collector.
	for _, oInfo := range oStream {
		// order rule 1: collection rule, collect the partial order info.
		updated = pab.collectPartials(oInfo)
	}

	if updated {
		// if the collector has been updated, try to process the committed partial orders.
		pab.processPartialOrder()
	}
}

// processPartialOrder is used to process partial order with phalanx anchor-based ordering rules.
func (pab *phalanxAnchorBasedOrdering) processPartialOrder() {
	for {
		// order rule 2: execution rule, select commands to execute with natural order.
		anchorSet := pab.fetchAnchorSet()

		// order rule 3: commitment rule, generate ordered blocks with free will.
		blocks, frontNo := pab.freeWill(anchorSet)
		if len(blocks) == 0 {
			// there isn't a committed inner block.
			break
		}

		// commit blocks.
		pab.logger.Debugf("[%d] commit front group, front-no. %d, safe %v, blocks count %d", pab.author, frontNo, anchorSet.Safe, len(blocks))
		for _, blk := range blocks {
			pab.seqNo++
			pab.exec.CommandExecution(blk, pab.seqNo)
			pab.reload.Committed(blk.Command.Author, blk.Command.Sequence)

			// record metrics.
			pab.metrics.CommitBlock(blk)
		}
	}
}

func (pab *phalanxAnchorBasedOrdering) collectPartials(oInfo types.OrderInfo) bool {
	// collect indicates the collection rule of phalanx:
	// which partial orders would be selected into execution process to compare order.
	pab.logger.Infof("[%d] collect partial order: %s", pab.author, oInfo.Format())

	// find the digest for current command the partial order refers to.
	commandD := oInfo.Command

	// check if current command has been committed or not.
	if pab.cRecorder.IsCommitted(commandD) {
		pab.logger.Debugf("[%d] committed command %s, ignore it", pab.author, commandD)
		return false
	}

	// push back partial order into recorder.queue.
	if err := pab.cRecorder.PushBack(oInfo); err != nil {
		pab.logger.Errorf("[%d] push back partial order failed: %s", pab.author, err)
		return false
	}

	// read command info from command cRecorder.
	info := pab.cRecorder.ReadCommandInfo(commandD)
	info.OrderAppend(oInfo)

	// already committed by quorum replicas, then update the timestamp list.
	if pab.cRecorder.IsQuorum(commandD) {
		info.UpdateTrustedTS(pab.oneCorrect)
	}

	// check the command status.
	switch info.OrderCount() {
	case pab.oneCorrect:
		// current command has reached correct sequenced status.
		pab.cRecorder.CorrectStatus(commandD)
		pab.logger.Infof("[%d] found correct sequenced command %s", pab.author, commandD)
	case pab.quorum:
		// current command has reached quorum sequenced status.
		pab.cRecorder.QuorumStatus(commandD)
		pab.logger.Infof("[%d] found quorum sequenced command %s", pab.author, commandD)
		info.UpdateTrustedTS(pab.oneCorrect)
	}
	return true
}

func (pab *phalanxAnchorBasedOrdering) fetchAnchorSet() types.FrontStream {
	// execute indicates the execution rule of phalanx:
	// which commands would be selected into commitment process to generate blocks.
	// here, we should take 'Natural Order' into thought.

	// oligarchy mode, relying on certain leader ordering.
	if pab.oligarchy != uint64(0) {
		return pab.oligarchyExecution()
	}

	// read the front set.
	commands, safe := pab.cRecorder.FrontCommands()

	var cStream types.CommandStream
	for _, digest := range commands {
		info := pab.cRecorder.ReadCommandInfo(digest)
		cStream = append(cStream, info)
	}

	if !safe {
		if qInfo := pab.cRecorder.PickQuorumInfo(); qInfo != nil {
			// we cannot make sure the validation of front set.
			cStream = interceptor.NewInterceptor(pab.author, pab.cRecorder, pab.oneCorrect, pab.logger).SelectToCommit(types.CommandStream{qInfo})
		}
	}

	return types.FrontStream{Safe: safe, Stream: cStream}
}

func (pab *phalanxAnchorBasedOrdering) oligarchyExecution() types.FrontStream {
	digest := pab.cRecorder.OligarchyLeaderFront(pab.oligarchy)
	commandInfo := pab.cRecorder.ReadCommandInfo(digest)
	if len(commandInfo.Orders) < pab.quorum {
		return types.FrontStream{Safe: true, Stream: nil}
	}
	return types.FrontStream{Safe: true, Stream: types.CommandStream{commandInfo}}
}

func (pab *phalanxAnchorBasedOrdering) freeWill(frontStream types.FrontStream) ([]types.InnerBlock, uint64) {
	// commit indicates the commitment rule of phalanx:
	// generate blocks and assign sequence order for them.
	// here, the block generation would follow the 'Free Will' of participants.
	if len(frontStream.Stream) == 0 {
		return nil, pab.frontNo
	}

	pab.frontNo++

	// free will:
	// generate blocks and sort according to the trusted timestamp
	// here, the command-pair with natural order cannot take part in concurrent command set.
	var sortable types.SortableInnerBlocks
	for _, frontC := range frontStream.Stream {
		// record metrics.
		pab.cMetrics.CommitFrontCommandInfo(frontC)

		// generate block, try to fetch the raw command to fulfill the block.
		rawCommand := pab.reader.ReadCommand(frontC.Digest)
		block := types.NewInnerBlock(pab.frontNo, frontStream.Safe, rawCommand, frontC.TrustedTS)
		pab.logger.Infof("[%d] generate block %s", pab.author, block.Format())

		// finished the block generation for command (digest), update the status of digest in command recorder.
		pab.cRecorder.CommittedStatus(frontC.Digest)

		// append the current block into sortable slice, waiting for order-determination.
		sortable = append(sortable, block)
	}

	// determine the order of commands which do not have any natural orders according to trusted timestamp.
	sort.Sort(sortable)

	return sortable, pab.frontNo
}
