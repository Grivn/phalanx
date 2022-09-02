package barrier

import "github.com/Grivn/phalanx/common/types"

func (barrier *commandStreamBarrierImpl) selection(unverifiedStream types.CommandStream) types.CommandStream {
	var additionalStream, returnStream types.CommandStream
	barrier.selected = make(map[string]bool)

	valid := true
	correctStream := barrier.cRecorder.ReadCSCInfos()
	quorumStream := barrier.cRecorder.ReadQSCInfos()

	for _, unverifiedC := range unverifiedStream {
		barrier.selected[unverifiedC.Digest] = true
	}

	returnStream = append(returnStream, unverifiedStream...)
	additionalStream, valid = barrier.filterStream(unverifiedStream, correctStream, quorumStream)

	if !valid {
		return nil
	}

	for {
		if len(additionalStream) == 0 {
			break
		}

		returnStream = append(returnStream, additionalStream...)
		additionalStream, valid = barrier.filterStream(additionalStream, correctStream, quorumStream)

		if !valid {
			return nil
		}
	}
	return returnStream
}

func (barrier *commandStreamBarrierImpl) filterStream(unverifiedStream, correctStream, quorumStream types.CommandStream) (types.CommandStream, bool) {
	var additionalStream types.CommandStream

	for _, unverifiedC := range unverifiedStream {
		barrier.selected[unverifiedC.Digest] = true
		pointer := make(map[uint64]uint64)

		for _, order := range unverifiedC.Orders {
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
				barrier.logger.Debugf("[%d] potential natural order (non-quorum): %s <- %s", barrier.author, correctC.Format(), unverifiedC.Format())
				return nil, false
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
				barrier.logger.Debugf("[%d] potential natural order (quorum): %s <- %s", barrier.author, quorumC.Format(), unverifiedC.Format())
				additionalStream = append(additionalStream, quorumC)
				barrier.selected[quorumC.Digest] = true
			}
		}
	}
	return additionalStream, true
}
