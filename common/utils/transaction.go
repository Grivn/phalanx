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
