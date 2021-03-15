package mocks

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"math/rand"
	"time"
)

func Shuffle(slice []*commonProto.OrderedMsg) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}
