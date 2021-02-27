package utils

import (
	"math/rand"
	"time"

	"github.com/Grivn/phalanx/common/types/protos"

	fTypes "github.com/ultramesh/flato-common/types"
)

func NewTx() *protos.Transaction {
	return &protos.Transaction{
		Tx: fTypes.NewTransaction(nil, nil, nil, time.Now().UnixNano(), int64(rand.Int())),
	}
}
