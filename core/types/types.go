package types

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type Config struct {
	N int

	Author uint64

	BatchSize int

	PoolSize int

	CommC commonTypes.CommChan

	ReliableC commonTypes.ReliableSendChan

	Executor external.Executor

	Network external.Network

	Logger external.Logger
}
