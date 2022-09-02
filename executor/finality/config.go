package finality

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
)

type Config struct {
	Author  uint64
	OLeader uint64
	N       int
	Mgr     api.MetaPool
	Manager api.TxManager
	Exec    external.ExecutionService
	Logger  external.Logger
	Metrics *metrics.Metrics
}
