package types

import (
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/api"
)

type Config struct {

}

type ReqConfig struct {
	Author  uint64
	Network external.Network
	SeqPool api.LogPool
	Auth    api.Authenticator
	Logger  external.Logger
}

type ColConfig struct {

}
