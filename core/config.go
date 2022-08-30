package phalanx

import (
	"time"

	"github.com/Grivn/phalanx/external"
)

type Config struct {
	Author      uint64
	OLeader     uint64
	Byz         bool
	OpenLatency int
	Duration    time.Duration
	Interval    int
	CDuration   time.Duration
	N           int
	Multi       int
	LogCount    int
	MemSize     int
	CommandSize int
	Selected    uint64
	Exec        external.ExecutionService
	Network     external.NetworkService
	Logger      external.Logger
}
