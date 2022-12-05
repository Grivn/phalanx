package memorypool

import (
	"github.com/Grivn/phalanx/external"
	"time"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/metrics"
)

type Config struct {
	Byz      bool
	Snapping bool
	Author   uint64
	N        int
	Multi    int
	Duration time.Duration
	Crypto   api.Crypto
	Logger   external.Logger
	Metrics  *metrics.MetaPoolMetrics
}
