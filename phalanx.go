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

type phalanxProvider struct {
	mgr internal.LogManager
	exe internal.Executor
	seq internal.SequencePool
}

func NewPhalanxProvider(n int, author uint64, exec external.Executor, network external.Network, logger external.Logger) *phalanxProvider {

	seq := sequencepool.NewSequencePool(n)

	exe := executor.NewExecutor(n, exec)

	mgr := logmanager.NewLogManager(n, author, seq, network, logger)

	return &phalanxProvider{
		mgr: mgr,
		seq: seq,
		exe: exe,
	}
}

func (phx *phalanxProvider) ProcessCommand(command *protos.Command) {
	if err := phx.mgr.ProcessCommand(command); err != nil {
		panic(err)
	}
}

func (phx *phalanxProvider) ProcessConsensusMessage(message *protos.ConsensusMessage) {
	switch message.Type {
	case protos.MessageType_PRE_ORDER:
		pre := &protos.PreOrder{}
		if err := proto.Unmarshal(message.Payload, pre); err != nil {
			panic(err)
		}
		if err := phx.mgr.ProcessPreOrder(pre); err != nil {
			panic(err)
		}
	case protos.MessageType_QUORUM_CERT:
		qc := &protos.QuorumCert{}
		if err := proto.Unmarshal(message.Payload, qc); err != nil {
			panic(err)
		}
		if err := phx.mgr.ProcessQC(qc); err != nil {
			panic(err)
		}
	case protos.MessageType_VOTE:
		vote := &protos.Vote{}
		if err := proto.Unmarshal(message.Payload, vote); err != nil {
			panic(err)
		}
		if err := phx.mgr.ProcessVote(vote); err != nil {
			panic(err)
		}
	}
}

func (phx *phalanxProvider) MakePayload() ([]byte, error) {
	return phx.seq.PullQCs()
}

func (phx *phalanxProvider) VerifyPayload(payload []byte) error {
	if err := phx.seq.VerifyQCs(payload); err != nil {
		return fmt.Errorf("phalanx verify failed: %s", err)
	}
	return nil
}

func (phx *phalanxProvider) CommitPayload(payload []byte) error {
	if err := phx.exe.CommitQCs(payload); err != nil {
		return fmt.Errorf("phalanx execution failed: %s", err)
	}
	return nil
}
