package logmgr

import (
	"github.com/Grivn/phalanx/api"
	"github.com/Grivn/phalanx/external"
)

type logMgrImpl struct {
	n int

	author uint64

	logpool map[uint64]*logPool

	auth api.Authenticator

	logger external.Logger
}

func newLogMgrImpl() *logMgrImpl {
	return &logMgrImpl{}
}
