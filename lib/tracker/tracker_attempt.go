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

	// partialMap records the partial orders which current node has received.
	attemptMap sync.Map

	// logger prints logs.
	logger external.Logger
}

func NewAttemptTracker(author uint64, logger external.Logger) api.AttemptTracker {
	logger.Infof("[%d] initiate command tracker", author)
	return &attemptTracker{
		author: author,
		logger: logger,
	}
}

// Record records order-attempt information.
func (at *attemptTracker) Record(attempt *protos.OrderAttempt) {}

// Get gets order-attempt according to query index.
func (at *attemptTracker) Get(idx types.QueryIndex) *protos.OrderAttempt { return nil }

// Checkpoint records the checkpoint and take garbage-collection according to it.
func (at *attemptTracker) Checkpoint(checkpoint *protos.Checkpoint) {}

// HasCheckpoint checks if current checkpoint exists.
func (at *attemptTracker) HasCheckpoint(idx types.QueryIndex) bool { return true }

// GarbageCollect collects garbage according to current status.
func (at *attemptTracker) GarbageCollect() {}
