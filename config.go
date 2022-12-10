package phalanx

import (
	"time"

	"github.com/Grivn/phalanx/pkg/external"
)

type Config struct {
	OligarchID  uint64
	IsByzantine bool
	IsSnapping  bool
	NodeID      uint64
	NodeCount   int
	OpenLatency int
	Duration    time.Duration
	Interval    int
	CDuration   time.Duration
	Multi       int
	LogCount    int
	MemSize     int
	CommandSize int
	Selected    uint64
	PrivateKey  external.PrivateKey
	PublicKeys  external.PublicKeys
	Executor    external.Executor
	Network     external.Sender
	Logger      external.Logger
}
