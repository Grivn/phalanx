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
}

func NewPhalanxProvider(n int, author uint64, exec external.ExecutorService, network external.NetworkService, logger external.Logger) *phalanxImpl {

	seq := sequencepool.NewSequencePool(author, n)

	exe := executor.NewExecutor(n, exec)

	mgr := logmanager.NewLogManager(n, author, seq, network, logger)

	return &phalanxImpl{
		logManager:   mgr,
		sequencePool: seq,
		executor:     exe,
	}
}

func (phi *phalanxImpl) ProcessCommand(command *protos.Command) {
	phi.sequencePool.InsertCommand(command)
	if err := phi.logManager.ProcessCommand(command); err != nil {
		panic(err)
	}
}

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

func (phi *phalanxImpl) MakePayload() ([]byte, error) {
	return phi.sequencePool.PullQCs()
}

func (phi *phalanxImpl) VerifyPayload(payload []byte) error {
	if err := phi.sequencePool.VerifyQCs(payload); err != nil {
		return fmt.Errorf("phalanx verify failed: %s", err)
	}
	return nil
}

func (phi *phalanxImpl) StablePayload(payload []byte) error {
	return phi.sequencePool.StableQCs(payload)
}

func (phi *phalanxImpl) CommitPayload(payload []byte) error {
	if err := phi.executor.CommitQCs(payload); err != nil {
		return fmt.Errorf("phalanx execution failed: %s", err)
	}
	return nil
}
