package phalanx

import (
	"fmt"
	"time"

	"github.com/Grivn/phalanx/common/crypto"
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
	logger       external.Logger
}

func NewPhalanxProvider(n int, author uint64, size int, duration time.Duration, exec external.ExecuteService, network external.NetworkService, logger external.Logger) *phalanxImpl {
	_ = crypto.SetKeys()

	seq := sequencepool.NewSequencePool(author, n, size, duration, logger)

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
		pOrder := &protos.PartialOrder{}
		if err := proto.Unmarshal(message.Payload, pOrder); err != nil {
			panic(err)
		}
		if err := phi.logManager.ProcessPartial(pOrder); err != nil {
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
	pBatch, err := phi.sequencePool.PullPartials()
	if err != nil {
		return nil, err
	}

	for _, pOrder := range pBatch.Partials {
		phi.logger.Infof("payload generation: replica %d sequence %d digest %s", pOrder.Author(), pOrder.Sequence(), pOrder.CommandDigest())
	}

	payload, err := marshal(pBatch)
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
	phi.sequencePool.RestorePartials()
}

// Verify is used to verify the phalanx payload.
// here, we would like to verify the validation of phalanx partial order, and record which seqNo has already been proposed.
func (phi *phalanxImpl) Verify(payload []byte) error {
	pBatch, err := unmarshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	err = phi.sequencePool.VerifyPartials(pBatch);
	if err != nil {
		return fmt.Errorf("phalanx verify failed: %s", err)
	}
	return nil
}

// SetStable is used to set stable
// here we would like to use it to control the order to process phalanx partial order.
// when such a interface has returned error, a timeout event should be triggered.
// 1) chained-bft: for each round we have generated a pBatch for chained-bft, we would like to use
//    it to set phalanx stable status.
// 2) classic-bft: for every time we are trying to execute a block, we would like to use it to
//    set phalanx stable status.
func (phi *phalanxImpl) SetStable(payload []byte) error {
	pBatch, err := unmarshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	return phi.sequencePool.SetStablePartials(pBatch)
}

// Commit is used to commit the phalanx partial order batch which has been verified by bft consensus.
func (phi *phalanxImpl) Commit(payload []byte) error {
	pBatch, err := unmarshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	err = phi.executor.CommitPartials(pBatch)
	if  err != nil {
		return fmt.Errorf("phalanx execution failed: %s", err)
	}
	return nil
}

func marshal(pBatch *protos.PartialOrderBatch) ([]byte, error) {
	return proto.Marshal(pBatch)
}

func unmarshal(payload []byte) (*protos.PartialOrderBatch, error) {
	pBatch := &protos.PartialOrderBatch{}
	if err := proto.Unmarshal(payload, pBatch); err != nil {
		return nil, fmt.Errorf("invalid partial order batch: %s", err)
	}

	return pBatch, nil
}
