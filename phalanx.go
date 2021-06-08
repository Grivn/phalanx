package phalanx

import (
	"github.com/Grivn/phalanx/external"
)

type phalanx struct {
	n int
	f int
	author uint64

	logger external.Logger
}

func newPhalanxImpl() *phalanx {
	return nil
}
