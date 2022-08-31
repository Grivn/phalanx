package finality

import (
	"sort"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
)

type orderMediumT struct {
	seqNo uint64

	// logger is used to print logs.
	logger external.Logger

	//
	metrics *metrics.OrderRuleMetrics

	blocks types.SortableInnerBlocks
}

func newOrderMediumT(conf Config) *orderMediumT {
	return &orderMediumT{
		blocks:  nil,
		logger:  conf.Logger,
		metrics: conf.Metrics.MediumTimeMetrics,
	}
}

func (ot *orderMediumT) commitAccordingMediumT(block types.InnerBlock) {
	ot.blocks = append(ot.blocks, block)

	if len(ot.blocks) < 1000 {
		return
	}
	sort.Sort(ot.blocks)
	commitBlocks := ot.blocks[:10]
	updateBlocks := types.SortableInnerBlocks{}
	for index, blk := range ot.blocks {
		if index < 10 {
			continue
		}
		updateBlocks = append(updateBlocks, blk)
	}
	ot.blocks = updateBlocks

	for _, blk := range commitBlocks {
		ot.metrics.CommitBlock(blk)
	}
}
