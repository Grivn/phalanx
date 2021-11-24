package execcomplex

import (
	"github.com/Grivn/phalanx/internal"
	"sort"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"

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

	// cRecorder is used to record the command info.
	cRecorder internal.CommandRecorder

	// democracy is used to generate block with free will committee.
	democracy map[uint64]*btree.BTree

	// reader is used to read raw commands from meta pool.
	reader internal.MetaReader

	// logger is used to print logs.
	logger external.Logger
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
		cRecorder:  recorder,
		democracy:  democracy,
		reader:     reader,
		logger:     logger,
	}
}

func (cr *commitmentRule) freeWill(executionInfos []*types.CommandInfo) []types.InnerBlock {
	if len(executionInfos) == 0 {
		return nil
	}

	// free will: init the democracy committee with raw data.
	for _, eInfo := range executionInfos {
		cr.logger.Debugf("[%d] execution info %s", cr.author, eInfo.Format())
		for _, pOrder := range eInfo.Orders {
			cr.democracy[pOrder.Author()].ReplaceOrInsert(pOrder)
			cr.logger.Debugf("[%d]    collected partial order %s", cr.author, pOrder.Format())
		}
	}

	// free will: trying to generate blocks.
	var blocks []types.InnerBlock
	for {
		concurrentC := cr.generateConcurrentC()
		sub := cr.generateSortedBlocks(concurrentC)
		blocks = append(blocks, sub...)

		if len(blocks) == len(executionInfos) {
			break
		}
	}

	return blocks
}

func (cr *commitmentRule) generateConcurrentC() []string {
	// free will:
	// we would like to find the first concurrent command set in current democracy committee,
	// and produce a slice for concurrent commands' digest for advanced processing.
	var concurrentC []string
	counter := make(map[string]int)

	// read the front command of partial order on each replica.
	for _, will := range cr.democracy {
		item := will.Min()

		if item == nil {
			continue
		}

		pOrder := item.(*protos.PartialOrder)
		counter[pOrder.CommandDigest()]++
	}

	// if there is at least one correct node (f+1) believing one specific command should be the front,
	// we should put it into concurrent slice. the one correct set could make sure that it is a preference from
	// correct node, and we would like to put it first.
	for digest, count := range counter {
		if count >= cr.oneCorrect {
			concurrentC = append(concurrentC, digest)
			cr.logger.Debugf("[%d] concurrent command set append %s", cr.author, digest)
		}
	}

	// if there is not any command selected into concurrent slice, it means we cannot find a correct set for
	// concurrent command, so that put all the command in the front of replica's partial order into the concurrent
	// slice.
	if len(concurrentC) == 0 {
		for digest := range counter {
			concurrentC = append(concurrentC, digest)
			cr.logger.Debugf("[%d] concurrent command set append %s", cr.author, digest)
		}
	}
	return concurrentC
}

func (cr *commitmentRule) generateSortedBlocks(concurrentC []string) []types.InnerBlock {
	// free will:
	// generate blocks and sort according to the trusted timestamp
	// here, the command-pair with natural order cannot take part in concurrent command set.
	var sortable types.SortableInnerBlocks
	for _, digest := range concurrentC {
		// read the command info from command recorder.
		info := cr.cRecorder.ReadCommandInfo(digest)

		// generate block, try to fetch the raw command to fulfill the block.
		rawCommand := cr.reader.ReadCommand(info.CurCmd)
		block := types.NewInnerBlock(rawCommand, info.Timestamps[cr.fault])
		cr.logger.Infof("[%d] generate block %s", cr.author, block.Format())

		// finished the block generation for command (digest), update the status of digest in command recorder.
		cr.cRecorder.CommittedStatus(info.CurCmd)

		// append the current block into sortable slice, waiting for order-determination.
		sortable = append(sortable, block)

		// remove the partial order from democracy committee.
		for _, pOrder := range info.Orders {
			cr.democracy[pOrder.Author()].Delete(pOrder)
		}
	}

	// determine the order of commands which do not have any natural orders according to trusted timestamp.
	sort.Sort(sortable)
	return sortable
}
