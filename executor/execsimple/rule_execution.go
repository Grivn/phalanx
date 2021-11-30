package execsimple

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

type executionRule struct {
	// author indicates the identifier of current node.
	author uint64

	// n indicates the number of replicas.
	n int

	// oneCorrect indicates there is at least one correct node for bft.
	oneCorrect int

	// quorum indicates the legal size for bft.
	quorum int

	// cRecorder is used to record the command info.
	cRecorder internal.CommandRecorder

	selected map[string]bool

	// logger is used to print logs.
	logger external.Logger
}

func newExecutionRule(author uint64, n int, recorder internal.CommandRecorder, logger external.Logger) *executionRule {
	logger.Infof("[%d] initiate natural order handler, replica count %d", author, n)
	return &executionRule{
		author:     author,
		n:          n,
		oneCorrect: types.CalculateOneCorrect(n),
		quorum:     types.CalculateQuorum(n),
		cRecorder:  recorder,
		logger:     logger,
	}
}

func (er *executionRule) execution() (cStream types.CommandStream) {

	// read the front set.
	commands, verified := er.cRecorder.FrontCommands()

	for _, digest := range commands {
		info := er.cRecorder.ReadCommandInfo(digest)
		cStream = append(cStream, info)
	}

	if !verified {
		// we cannot make sure the validation of front set.
		cStream = er.selection(cStream)
	}

	return cStream
}

func (er *executionRule) selection(unverifiedStream types.CommandStream) types.CommandStream {
	correctStream := er.cRecorder.ReadCSCInfos()

	quorumStream := er.cRecorder.ReadQSCInfos()

	er.selected = make(map[string]bool)

	var additionalStream types.CommandStream

	var returnStream types.CommandStream

	valid := true

	for _, unverifiedC := range unverifiedStream {
		er.selected[unverifiedC.CurCmd] = true
	}

	returnStream = append(returnStream, unverifiedStream...)
	additionalStream, valid = er.filterStream(unverifiedStream, correctStream, quorumStream)

	if !valid {
		return nil
	}

	for {
		if len(additionalStream) == 0 {
			break
		}

		returnStream = append(returnStream, additionalStream...)
		additionalStream, valid = er.filterStream(additionalStream, correctStream, quorumStream)

		if !valid {
			return nil
		}
	}
	return returnStream
}

func (er *executionRule) filterStream(unverifiedStream, correctStream, quorumStream types.CommandStream) (types.CommandStream, bool) {
	var additionalStream types.CommandStream

	for _, unverifiedC := range unverifiedStream {
		er.selected[unverifiedC.CurCmd] = true
		pointer := make(map[uint64]uint64)

		for _, order := range unverifiedC.Orders {
			pointer[order.Author] = order.Sequence
		}

		for _, correctC := range correctStream {
			count := 0

			for id, seq := range pointer {
				oInfo, ok := correctC.Orders[id]
				if !ok || oInfo.Sequence < seq {
					count++
				}
				if count == er.oneCorrect {
					break
				}
			}

			if count < er.oneCorrect {
				er.logger.Debugf("[%d] potential natural order (non-quorum): %s <- %s", er.author, correctC.Format(), unverifiedC.Format())
				return nil, false
			}
		}

		for _, quorumC := range quorumStream {
			if er.selected[quorumC.CurCmd] {
				continue
			}

			count := 0

			for id, seq := range pointer {
				oInfo, ok := quorumC.Orders[id]
				if !ok || oInfo.Sequence < seq {
					count++
				}
				if count == er.oneCorrect {
					break
				}
			}

			if count < er.oneCorrect {
				er.logger.Debugf("[%d] potential natural order (quorum): %s <- %s", er.author, quorumC.Format(), unverifiedC.Format())
				additionalStream = append(additionalStream, quorumC)
				er.selected[quorumC.CurCmd] = true
			}
		}
	}
	return additionalStream, true
}
