package conengine

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/external"
)

type Config struct {
	NodeID   uint64
	N        int
	Crypto   api.Crypto
	External external.External
}
