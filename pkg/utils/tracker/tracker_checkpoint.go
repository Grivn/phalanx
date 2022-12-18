package tracker

import (
	"sync"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
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

func (ct *checkpointTracker) Record(checkpoint *protos.Checkpoint) {
	qIdx := types.QueryIndex{Author: checkpoint.NodeID(), SeqNo: checkpoint.SeqNo()}

	if _, ok := ct.checkpointMap.Load(qIdx); ok {
		ct.logger.Debugf("[%d] duplicated checkpoint %s", ct.author, checkpoint.Format())
		return
	}

	ct.logger.Debugf("[%d] record checkpoint %s", ct.author, checkpoint.Format())
	ct.checkpointMap.Store(qIdx, checkpoint)
}

func (ct *checkpointTracker) Get(idx types.QueryIndex) *protos.Checkpoint {
	e, ok := ct.checkpointMap.Load(idx)
	if !ok {
		return nil
	}
	checkpoint := e.(*protos.Checkpoint)
	ct.checkpointMap.Delete(idx)
	return checkpoint
}

func (ct *checkpointTracker) IsExist(idx types.QueryIndex) bool {
	_, ok := ct.checkpointMap.Load(idx)
	return ok
}
