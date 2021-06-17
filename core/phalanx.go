package phalanx

import (
	"fmt"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/executor"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/logmanager"
	"github.com/Grivn/phalanx/sequencepool"

	"github.com/gogo/protobuf/proto"
)

type phalanxImpl struct {
	logManager   internal.LogManager
	executor     internal.Executor
	sequencePool internal.SequencePool

	logger external.Logger
}

func NewPhalanxProvider(n int, author uint64, exec external.ExecutorService, network external.NetworkService, logger external.Logger) *phalanxImpl {

	seq := sequencepool.NewSequencePool(author, n)

	exe := executor.NewExecutor(n, exec)

	mgr := logmanager.NewLogManager(n, author, seq, network, logger)

	return &phalanxImpl{
		logManager:   mgr,
		sequencePool: seq,
		executor:     exe,
		logger:       logger,
	}
}

// ProcessCommand is used to process the commands from clients.
func (phi *phalanxImpl) ProcessCommand(command *protos.Command) {
	phi.sequencePool.InsertCommand(command)
	if err := phi.logManager.ProcessCommand(command); err != nil {
		panic(err)
	}
}

// ProcessConsensusMessage is used process the consensus messages from phalanx replica.
func (phi *phalanxImpl) ProcessConsensusMessage(message *protos.ConsensusMessage) {
	switch message.Type {
	case protos.MessageType_PRE_ORDER:
		pre := &protos.PreOrder{}
		if err := proto.Unmarshal(message.Payload, pre); err != nil {
			panic(err)
		}
		if err := phi.logManager.ProcessPreOrder(pre); err != nil {
			panic(err)
		}
	case protos.MessageType_QUORUM_CERT:
		qc := &protos.QuorumCert{}
		if err := proto.Unmarshal(message.Payload, qc); err != nil {
			panic(err)
		}
		if err := phi.logManager.ProcessQC(qc); err != nil {
			panic(err)
		}
	case protos.MessageType_VOTE:
		vote := &protos.Vote{}
		if err := proto.Unmarshal(message.Payload, vote); err != nil {
			panic(err)
		}
		if err := phi.logManager.ProcessVote(vote); err != nil {
			panic(err)
		}
	}
}

// MakePayload is used to generate payloads for bft consensus.
func (phi *phalanxImpl) MakePayload() ([]byte, error) {
	qcb, err := phi.sequencePool.PullQCs()
	if err != nil {
		return nil, err
	}

	for _, qc := range qcb.Filters[0].QCs {
		phi.logger.Infof("payload generation: replica %d sequence %d digest %s", qc.Author(), qc.Sequence(), qc.CommandDigest())
	}

	payload, err := marshal(qcb)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (phi *phalanxImpl) BecomeLeader() {
	phi.sequencePool.BecomeLeader()
}

// Restore is used to restore data when we have found a timeout event in partial-synchronized bft consensus module.
func (phi *phalanxImpl) Restore() {
	phi.sequencePool.RestoreQCs()
}

// Verify is used to verify the phalanx payload.
// here, we would like to verify the validation of phalanx QCs, and record which seqNo has already been proposed.
func (phi *phalanxImpl) Verify(payload []byte) error {
	qcb, err := unmarshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	err = phi.sequencePool.VerifyQCs(qcb);
	if err != nil {
		return fmt.Errorf("phalanx verify failed: %s", err)
	}
	return nil
}

// SetStable is used to set stable
// here we would like to use it to control the order to process phalanx QCs.
// when such a interface has returned error, a timeout event should be triggered.
// 1) chained-bft: for each round we have generated a QC for chained-bft, we would like to use
//    it to set phalanx stable status.
// 2) classic-bft: for every time we are trying to execute a block, we would like to use it to
//    set phalanx stable status.
func (phi *phalanxImpl) SetStable(payload []byte) error {
	qcb, err := unmarshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	return phi.sequencePool.SetStableQCs(qcb)
}

// Commit is used to commit the phalanx-QCBatch which has been verified by bft consensus.
func (phi *phalanxImpl) Commit(payload []byte) error {
	qcb, err := unmarshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	err = phi.executor.CommitQCs(qcb)
	if  err != nil {
		return fmt.Errorf("phalanx execution failed: %s", err)
	}
	return nil
}

func marshal(qcb *protos.QCBatch) ([]byte, error) {
	return proto.Marshal(qcb)
}

func unmarshal(payload []byte) (*protos.QCBatch, error) {
	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return nil, fmt.Errorf("invalid QC-batch: %s", err)
	}

	return qcb, nil
}
