package utils

import (
	"math/rand"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/common/types/protos"
)

func NewTx() *protos.Transaction {
	payload := make([]byte, 1024)
	rand.Read(payload)
	return types.GenerateTransaction(payload)
}

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
