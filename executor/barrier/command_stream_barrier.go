package barrier

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type commandStreamBarrierImpl struct {
	// author indicates the identifier of current node.
	author uint64

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// cRecorder is used to record the command info.
	cRecorder api.CommandRecorder

	// selected is referred to the commands to be committed.
	selected map[string]bool

	// logger is used to print logs.
	logger external.Logger
}

func NewCommandStreamBarrier(author uint64, cRecorder api.CommandRecorder, oneCorrect int, logger external.Logger) *commandStreamBarrierImpl {
	return &commandStreamBarrierImpl{
		author:     author,
		oneCorrect: oneCorrect,
		cRecorder:  cRecorder,
		selected:   make(map[string]bool),
		logger:     logger,
	}
}

func (barrier *commandStreamBarrierImpl) BaselineGroup(baselines types.CommandStream) types.CommandStream {
	return barrier.processRiskPath(baselines)
}

func (barrier *commandStreamBarrierImpl) processRiskPath(baselines types.CommandStream) types.CommandStream {
	var frontStream types.CommandStream
	barrier.selected = make(map[string]bool)

	for _, baseline := range baselines {
		barrier.selected[baseline.Digest] = true
	}

	correctStream := barrier.cRecorder.ReadCSCInfos()
	quorumStream := barrier.cRecorder.ReadQSCInfos()

	for _, baseline := range baselines {
		pointer := make(map[uint64]uint64)
		for _, order := range baseline.Orders {
			pointer[order.Author] = order.Sequence
		}

		for _, correctC := range correctStream {
			count := 0

			for id, seq := range pointer {
				oInfo, ok := correctC.Orders[id]
				if !ok || oInfo.Sequence > seq {
					count++
				}
				if count == barrier.oneCorrect {
					break
				}
			}

			if count < barrier.oneCorrect {
				barrier.logger.Debugf("[%d] potential natural order (non-quorum): %s <- %s", barrier.author, correctC.Format(), baseline.Format())
				return nil
			}
		}

		for _, quorumC := range quorumStream {
			if barrier.selected[quorumC.Digest] {
				continue
			}

			count := 0

			for id, seq := range pointer {
				oInfo, ok := quorumC.Orders[id]
				if !ok || oInfo.Sequence > seq {
					count++
				}
				if count == barrier.oneCorrect {
					break
				}
			}

			if count < barrier.oneCorrect {
				barrier.logger.Debugf("[%d] potential natural order (quorum): %s <- %s", barrier.author, quorumC.Format(), baseline.Format())
				frontStream = append(frontStream, quorumC)
				barrier.selected[quorumC.Digest] = true
			}
		}
	}

	return append(baselines, frontStream...)
}
