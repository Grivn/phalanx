package tracker

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sync"
)

type attemptTracker struct {
	// author indicates current node identifier.
	author uint64

	// partialMap records the order-attempts current node has received.
	attemptMap sync.Map

	// logger prints logs.
	logger external.Logger
}

func NewAttemptTracker(author uint64, logger external.Logger) api.AttemptTracker {
	logger.Infof("[%d] initiate attempt tracker", author)
	return &attemptTracker{
		author: author,
		logger: logger,
	}
}

func (at *attemptTracker) Record(attempt *protos.OrderAttempt) {}

func (at *attemptTracker) Get(idx types.QueryIndex) *protos.OrderAttempt { return nil }
