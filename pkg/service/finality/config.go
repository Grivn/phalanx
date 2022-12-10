package finality

import (
	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/external"
)

type Config struct {
	Author  uint64
	OLeader uint64
	N       int
	Pool    api.MetaPool
	Exec    external.Executor
	Logger  external.Logger
	Metrics *metrics.Metrics
}
