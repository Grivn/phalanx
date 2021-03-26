package binsubset

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type recorderMgr struct {
	n int
	f int

	finished bool

	sequence uint64

	author uint64

	tags map[uint64]*commonProto.BinaryTag

	qSet []byte

	logger external.Logger
}

func newRecorder(n, f int, author, sequence uint64, logger external.Logger) *recorderMgr {
	return &recorderMgr{
		n: n,
		f: f,
		author: author,
		sequence: sequence,
		tags: make(map[uint64]*commonProto.BinaryTag),
		logger: logger,
	}
}

func (re *recorderMgr) record(author uint64, tag *commonProto.BinaryTag) *commonProto.BinaryTag {
	re.logger.Infof("replica %d received notification for sequence %d from replica %d, tag %v", re.author, re.sequence, author, tag.BinarySet)
	re.tags[author] = tag

	_, ok := re.tags[re.author]
	if !ok {
		return nil
	}

	if len(re.tags) < re.quorum() {
		return nil
	}

	re.generate()
	re.logger.Debugf("replica %d generate local quorum set for sequence %d, set %v", re.author, re.sequence, re.qSet)

	if re.isSupervisor() && !re.finished {
		re.finished = true
		re.logger.Infof("replica %d is the supervisor of sequence %d, send quorum event, set %v", re.author, re.sequence, re.qSet)
		return &commonProto.BinaryTag{
			Sequence:   re.sequence,
			BinarySet:  re.qSet,
			BinaryHash: commonTypes.CalculatePayloadHash(re.qSet, 0),
		}
	}
	return nil
}

func (re *recorderMgr) generate() {
	counter := make(map[int]int)
	set := make([]byte, re.n)
	for _, tag := range re.tags {
		for index, val := range tag.BinarySet {
			if val == 1 {
				counter[index]++
			}

			if counter[index] >= re.oneQuorum() {
				set[index] = 1
			}
		}
	}
	re.qSet = set
}

func (re *recorderMgr) compare(set []byte) bool {
	if len(set) == 0 {
		re.logger.Warningf("replica %d received a nil quorum set", re.author)
		return false
	}

	if len(re.qSet) == 0 {
		re.logger.Debugf("replica %d quorum set is nil", re.author)
		return false
	}

	re.logger.Infof("replica %d compare the q-set of remote and local, remote %v local %v", re.author, set, re.qSet)

	for index := range set {
		if set[index] == 1 && re.qSet[index] == 0 {
			re.logger.Debugf("replica %d miss the value on replica %d", re.author, uint64(index+1))
			return false
		}
	}

	re.logger.Debugf("replica %d pass the comparison", re.author)
	return true
}

func (re *recorderMgr) quorum() int {
	return re.n - re.f
}

func (re *recorderMgr) oneQuorum() int {
	return re.f + 1
}

func (re *recorderMgr) isSupervisor() bool {
	return re.author == ((re.sequence-1)%uint64(re.n)+1)
}
