package receiver

import "github.com/Grivn/phalanx/external"

type Config struct {
	Auction     bool
	Author      uint64
	Multi       int
	CommandSize int
	MemSize     int
	Selected    uint64
	Sender      external.NetworkService
	Logger      external.Logger
}
