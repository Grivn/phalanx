package sequencer

import (
	"time"

	"github.com/Grivn/phalanx/pkg/external"
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
	Sender      external.Sender
	Logger      external.Logger
}
