package binbyzantine

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type binary struct {
	n int

	author uint64

	hash string

	sequence uint64

	bits map[uint64]bool

	tag *commonProto.BinaryTag

	logger external.Logger
}

func newBinary(n int, author uint64, tag *commonProto.BinaryTag, logger external.Logger) *binary {
	logger.Infof("replica %d init binary instance for sequence %d, bit set %v", author, tag.Sequence, tag.BinarySet)

	bits := make(map[uint64]bool)

	for index, value := range tag.BinarySet {
		if value == 1 {
			bits[uint64(index+1)] = true
		}
	}

	return &binary{
		n:        n,
		author:   author,
		hash:     tag.BinaryHash,
		sequence: tag.Sequence,
		bits:     bits,
		tag:      tag,
		logger:   logger,
	}
}

func (binary *binary) compare(set []byte) bool {
	changed := false
	for index, value := range set {
		id := uint64(index+1)
		if value == 1 && !binary.bits[id] {
			binary.logger.Infof("replica %d turn on the bit of %d for sequence %d", binary.author, id, binary.sequence)
			binary.bits[id] = true
			changed = true
		}
	}
	return changed
}

func (binary *binary) include(tag *commonProto.BinaryTag) bool {
	for index, value := range tag.BinarySet {
		id := uint64(index+1)
		if value == 1 && !binary.bits[id] {
			return false
		}
	}
	return true
}

func (binary *binary) convert() *commonProto.BinaryTag {
	set := make([]byte, binary.n)
	for index := range set {
		if binary.bits[uint64(index+1)] {
			set[index] = 1
		}
	}
	hash := commonTypes.CalculatePayloadHash(set, 0)

	bTag := &commonProto.BinaryTag{
		Sequence:   binary.sequence,
		BinaryHash: hash,
		BinarySet:  set,
	}
	return bTag
}