package mocks

import (
	"math/rand"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/common/types/protos"
)

func NewOrderedMessages(author uint64, from, to , typ protos.OrderType) []*protos.OrderedMsg {
	var list []*protos.OrderedMsg
	for index:= from; index<= to; index++ {
		payload := make([]byte, 1024)
		rand.Read(payload)

		bid := &protos.BatchId{
			Author:    author,
			BatchHash: types.CalculatePayloadHash(payload, 0),
		}

		msg := &protos.OrderedMsg{
			Type:     typ,
			Author:   author,
			Sequence: uint64(index),
			BatchId:  bid,
		}

		list = append(list, msg)
	}
	return list
}
