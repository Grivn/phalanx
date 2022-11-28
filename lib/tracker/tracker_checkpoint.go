package tracker

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sync"
)

type checkpointTracker struct {
	// author indicates current node identifier.
	author uint64

	// checkpointMap records the checkpoints current node has received.
	checkpointMap sync.Map

	// logger prints logs.
	logger external.Logger
}

func NewCheckpointTracker(author uint64, logger external.Logger) api.CheckpointTracker {
	logger.Infof("[%d] initiate checkpoint tracker", author)
	return &checkpointTracker{
		author: author,
		logger: logger,
	}
}

func (ct *checkpointTracker) Record(checkpoint *protos.Checkpoint) {}

func (ct *checkpointTracker) Get(idx types.QueryIndex) *protos.Checkpoint { return nil }

func (ct *checkpointTracker) IsExist(idx types.QueryIndex) bool { return true }
