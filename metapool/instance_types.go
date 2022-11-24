package metapool

import "github.com/Grivn/phalanx/common/protos"

type CheckpointEvent struct {
	Attempt   *protos.OrderAttempt
	QC        *protos.QuorumCert
	CallbackC chan *protos.Checkpoint
}

func (ev *CheckpointEvent) ToCheckpoint() *protos.Checkpoint {
	return &protos.Checkpoint{OrderAttempt: ev.Attempt, QC: ev.QC}
}
