package interceptor

import "github.com/Grivn/phalanx/common/types"

func (i *interceptorImpl) selectionOptional(barrier types.CommandStream) types.CommandStream {
	var additionalStream, returnStream types.CommandStream
	i.selected = make(map[string]bool)

	valid := true
	correctStream := i.cRecorder.ReadCSCInfos()
	quorumStream := i.cRecorder.ReadQSCInfos()

	for _, barrierC := range barrier {
		i.selected[barrierC.Digest] = true
	}

	returnStream = append(returnStream, barrier...)
	additionalStream, valid = i.filterStream(barrier, correctStream, quorumStream)

	if !valid {
		return nil
	}

	for {
		if len(additionalStream) == 0 {
			break
		}

		returnStream = append(returnStream, additionalStream...)
		additionalStream, valid = i.filterStream(additionalStream, correctStream, quorumStream)

		if !valid {
			return nil
		}
	}
	return returnStream
}

func (i *interceptorImpl) filterStream(barrier, correctStream, quorumStream types.CommandStream) (types.CommandStream, bool) {
	var additionalStream types.CommandStream

	for _, barrierC := range barrier {
		i.selected[barrierC.Digest] = true
		pointer := make(map[uint64]uint64)

		for _, order := range barrierC.Orders {
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
				i.logger.Debugf("[%d] potential natural order (non-quorum): %s <- %s", i.author, correctC.Format(), barrierC.Format())
				return nil, false
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
				i.logger.Debugf("[%d] potential natural order (quorum): %s <- %s", i.author, quorumC.Format(), barrierC.Format())
				additionalStream = append(additionalStream, quorumC)
				i.selected[quorumC.Digest] = true
			}
		}
	}
	return additionalStream, true
}
