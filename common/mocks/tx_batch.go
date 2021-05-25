package mocks

import (
	"math/rand"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

func NewTx() *protos.Transaction {
	payload := make([]byte, 1024)
	rand.Read(payload)
	return types.GenerateTransaction(payload)
}

func NewBatchId(author uint64, count int) []*protos.BatchId {
	var list []*protos.BatchId
	for index:=0; index<count; index++ {
		payload := make([]byte, 1024)
		rand.Read(payload)

		bid := &protos.BatchId{
			Author:    author,
			BatchHash: types.CalculatePayloadHash(payload, 0),
		}

		list = append(list, bid)
	}
	return list
}
