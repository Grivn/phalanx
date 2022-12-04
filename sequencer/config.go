package sequencer

import (
	"time"

	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
)

type Config struct {
	Byz         bool
	Snapping    bool
	Author      uint64
	N           int
	Multi       int
	CommandSize int
	MemSize     int
	Selected    uint64
	Duration    time.Duration
	Sender      external.NetworkService
	Logger      external.Logger
	Metrics     *metrics.MetaPoolMetrics
}
