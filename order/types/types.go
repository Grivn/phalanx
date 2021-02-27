package types

import (
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

type Config struct {

}

type ReqConfig struct {
	Author  uint64
	Network external.Network
	USIG    internal.USIG
	SeqPool internal.SequencePool
	Logger  external.Logger
}

type ColConfig struct {

}
