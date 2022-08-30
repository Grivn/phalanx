package finality

import (
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/metrics"
)

type Config struct {
	Author  uint64
	OLeader uint64
	N       int
	Mgr     internal.MetaPool
	Manager internal.TxManager
	Exec    external.ExecutionService
	Logger  external.Logger
	Metrics *metrics.Metrics
}
