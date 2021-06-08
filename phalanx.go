package phalanx

import (
	"github.com/Grivn/phalanx/external"
)

type phalanx struct {
	n int
	f int
	author uint64

	// logger is used to print logs
	logger external.Logger
}

func NewPhalanx() *phalanx {
	return nil
}
