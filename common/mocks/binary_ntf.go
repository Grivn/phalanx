package mocks

import (
	"math/rand"
	"time"

	"github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

func NewBinaryTag(n int, author, from, to uint64) []*commonProto.BinaryTag {
	f := (n - 1)/4
	//r := rand.Int()%f

	// random set size
	size := n - f

	r := rand.New(rand.NewSource(time.Now().Unix()))

	var tags []*commonProto.BinaryTag
	for seq:=from; seq<=to; seq++ {
		set := make([]byte, n)
		for i := range set {
			if i < size {
				set[i] = 1
			}
		}
		shuffle(r, set)
		set[int(author-1)]=1

		tag := &commonProto.BinaryTag{
			Sequence:   seq,
			BinarySet:  set,
			BinaryHash: types.CalculatePayloadHash(set, 0),
		}
		tags = append(tags, tag)
	}

	return tags
}

func shuffle(r *rand.Rand, slice []byte) {
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}
