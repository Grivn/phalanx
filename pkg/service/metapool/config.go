package metapool

import (
	"time"

	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/external"
)

type Config struct {
	Byz      bool
	Snapping bool
	Author   uint64
	N        int
	Multi    int
	Duration time.Duration
	Crypto   api.CryptoService
	Sender   external.Sender
	Logger   external.Logger
	Metrics  *metrics.MetaPoolMetrics
}
