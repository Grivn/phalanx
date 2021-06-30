package executor

import (
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

	// recorder is used to record the command info.
	recorder *commandRecorder

	// filter is used to generate block according to free will.
	filter map[uint64]*btree.BTree

	// logger is used to print logs.
	logger external.Logger
}

func newCommitmentRule(author uint64, n int, recorder *commandRecorder, logger external.Logger) *commitmentRule {
	filter := make(map[uint64]*btree.BTree)
	for i:=0; i<n; i++ {
		filter[uint64(i+1)] = btree.New(2)
	}

	return &commitmentRule{
		author:     author,
		n:          n,
		fault:      types.CalculateFault(n),
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		recorder:	recorder,
		filter:     filter,
		logger:     logger,
	}
}

func (cr *commitmentRule) freeWill(executionInfos []*commandInfo) []types.Block {
	if len(executionInfos) == 0 {
		return nil
	}

	var blocks []types.Block

	// init the raw data for free will.
	for _, eInfo := range executionInfos {
		for _, pOrder := range eInfo.pOrders {
			cr.filter[pOrder.Author()].ReplaceOrInsert(pOrder)
		}
	}

	for {
		var concurrentC []string
		counter := make(map[string]int)

		for _, will := range cr.filter {
			item := will.Min()

			if item == nil {
				continue
			}

			pOrder := item.(*protos.PartialOrder)
			counter[pOrder.CommandDigest()]++
		}

		for digest, count := range counter {
			if count >= cr.oneCorrect {
				concurrentC = append(concurrentC, digest)
			}
		}

		if len(concurrentC) == 0 {
			for digest := range counter {
				concurrentC = append(concurrentC, digest)
			}
		}

		sub := cr.generateSortedBlocks(concurrentC)
		blocks = append(blocks, sub...)

		if len(blocks) == len(executionInfos) {
			break
		}
	}

	return blocks
}

func (cr *commitmentRule) generateSortedBlocks(concurrentC []string) []types.Block {
	// generate blocks and sort according to the trusted timestamp
	// here, the command-pair with natural order cannot take part in concurrent command set.
	var sub types.SubBlock
	for _, digest := range concurrentC {
		info := cr.recorder.readCommandInfo(digest)
		block := types.NewBlock(info.curCmd, nil, nil, info.timestamps[cr.fault])
		rawCommand := cr.recorder.readCommandRaw(info.curCmd)
		if rawCommand != nil {
			block.TxList = rawCommand.Content
			block.HashList = rawCommand.HashList
		}
		cr.recorder.committedStatus(info.curCmd)
		sub = append(sub, block)

		// remove the partial order from filter b-trees.
		for _, pOrder := range info.pOrders {
			cr.filter[pOrder.Author()].Delete(pOrder)
		}
	}
	sort.Sort(sub)
	return sub
}
