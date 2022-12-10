package interceptor

import (
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
)

type interceptorImpl struct {
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

func NewInterceptor(author uint64, cRecorder api.CommandRecorder, oneCorrect int, logger external.Logger) api.Interceptor {
	return &interceptorImpl{
		author:     author,
		oneCorrect: oneCorrect,
		cRecorder:  cRecorder,
		selected:   make(map[string]bool),
		logger:     logger,
	}
}

func (i *interceptorImpl) SelectToCommit(barrier types.CommandStream) types.CommandStream {
	return i.selection(barrier)
}

func (i *interceptorImpl) selection(barrier types.CommandStream) types.CommandStream {
	var frontStream types.CommandStream
	i.selected = make(map[string]bool)

	for _, bInfo := range barrier {
		i.selected[bInfo.Digest] = true
	}

	correctStream := i.cRecorder.ReadCSCInfos()
	quorumStream := i.cRecorder.ReadQSCInfos()

	for _, bInfo := range barrier {
		pointer := make(map[uint64]uint64)
		for _, order := range bInfo.Orders {
			pointer[order.Author] = order.Sequence
		}

		for _, correctC := range correctStream {
			count := 0

			for id, seq := range pointer {
				oInfo, ok := correctC.Orders[id]
				if !ok || oInfo.Sequence > seq {
					count++
				}
				if count == i.oneCorrect {
					break
				}
			}

			if count < i.oneCorrect {
				i.logger.Debugf("[%d] potential natural order (non-quorum): %s <- %s", i.author, correctC.Format(), bInfo.Format())
				return nil
			}
		}

		for _, quorumC := range quorumStream {
			if i.selected[quorumC.Digest] {
				continue
			}

			count := 0

			for id, seq := range pointer {
				oInfo, ok := quorumC.Orders[id]
				if !ok || oInfo.Sequence > seq {
					count++
				}
				if count == i.oneCorrect {
					break
				}
			}

			if count < i.oneCorrect {
				i.logger.Debugf("[%d] potential natural order (quorum): %s <- %s", i.author, quorumC.Format(), bInfo.Format())
				frontStream = append(frontStream, quorumC)
				i.selected[quorumC.Digest] = true
			}
		}
	}

	return append(barrier, frontStream...)
}
