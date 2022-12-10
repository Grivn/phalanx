package conengine

import (
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/external"
)

type Config struct {
	NodeID   uint64
	N        int
	Crypto   api.CryptoService
	External external.External
}
