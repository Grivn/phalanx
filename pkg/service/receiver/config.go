package receiver

import (
	"github.com/Grivn/phalanx/pkg/external"
)

type Config struct {
	Auction     bool
	Author      uint64
	Multi       int
	CommandSize int
	MemSize     int
	Selected    uint64
	Sender      external.Sender
	Logger      external.Logger
}
