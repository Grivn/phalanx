package mocks

import (
	"math/rand"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/common/types/protos"
)

func NewOrderedMessages(author, from, to uint64) []*protos.OrderedReq {
	var list []*protos.OrderedReq
	for index:= from; index<= to; index++ {
		payload := make([]byte, 1024)
		rand.Read(payload)

		bid := &protos.BatchId{
			Author:    author,
			BatchHash: types.CalculatePayloadHash(payload, 0),
		}

		msg := &protos.OrderedReq{
			Author:   author,
			Sequence: index,
			BatchId:  bid,
		}

		list = append(list, msg)
	}
	return list
}
