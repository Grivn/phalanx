package execsimple

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"sort"
	"time"

	"github.com/google/btree"
)

type commitmentRule struct {
	// author indicates the identifier of current node.
	author uint64

	// n indicates the number of replicas.
	n int

	// fault indicates the max amount byzantine node in current cluster.
	fault int

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// frontNo is used to track the sequence number for front stream.
	frontNo uint64

	// cRecorder is used to record the command info.
	cRecorder internal.CommandRecorder

	// democracy is used to generate block with free will committee.
	democracy map[uint64]*btree.BTree

	// reader is used to read raw commands from meta pool.
	reader internal.MetaReader

	// logger is used to print logs.
	logger external.Logger

	// totalCommandInfo is the number of committed command info.
	totalCommandInfo int

	//
	intervalCommandInfo int

	// totalLatency is the total latency since command info generation to be committed.
	totalLatency int64

	//
	intervalLatency int64
}

func newCommitmentRule(author uint64, n int, recorder internal.CommandRecorder, reader internal.MetaReader, logger external.Logger) *commitmentRule {
	logger.Infof("[%d] initiate free will committee, replica count %d", author, n)
	democracy := make(map[uint64]*btree.BTree)
	for i:=0; i<n; i++ {
		democracy[uint64(i+1)] = btree.New(2)
	}

	return &commitmentRule{
		author:     author,
		n:          n,
		fault:      types.CalculateFault(n),
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		frontNo:    uint64(0),
		cRecorder:  recorder,
		democracy:  democracy,
		reader:     reader,
		logger:     logger,
	}
}

func (cr *commitmentRule) freeWill(frontStream types.FrontStream) ([]types.InnerBlock, uint64) {
	if len(frontStream.Stream) == 0 {
		return nil, cr.frontNo
	}

	cr.frontNo++

	// free will:
	// generate blocks and sort according to the trusted timestamp
	// here, the command-pair with natural order cannot take part in concurrent command set.
	var sortable types.SortableInnerBlocks
	for _, frontC := range frontStream.Stream {
		sub := time.Now().UnixNano() - frontC.GTime
		cr.totalCommandInfo++
		cr.totalLatency += sub
		cr.intervalCommandInfo++
		cr.intervalLatency += sub

		// generate block, try to fetch the raw command to fulfill the block.
		rawCommand := cr.reader.ReadCommand(frontC.CurCmd)
		block := types.NewInnerBlock(cr.frontNo, frontStream.Safe, rawCommand, frontC.Timestamps[cr.fault])
		cr.logger.Infof("[%d] generate block %s", cr.author, block.Format())

		// finished the block generation for command (digest), update the status of digest in command recorder.
		cr.cRecorder.CommittedStatus(frontC.CurCmd)

		// append the current block into sortable slice, waiting for order-determination.
		sortable = append(sortable, block)
	}

	// determine the order of commands which do not have any natural orders according to trusted timestamp.
	sort.Sort(sortable)

	return sortable, cr.frontNo
}
