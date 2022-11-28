package consensus

import (
	"fmt"
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/gogo/protobuf/proto"
)

type consensusRunner struct {
	// nodeID is the identifier of current node.
	nodeID uint64

	// n is the total number of the participants.
	n int

	// quorum is the threshold of stable consensus.
	quorum int

	// aggMap is used to create aggregated-QC for checkpoints.
	aggMap map[string]*protos.Checkpoint

	// consensusMessageC is used to relay the consensus message into consensus runner.
	consensusMessageC chan *protos.ConsensusMessage

	// closeC is used to stop the processor of consensus runner.
	closeC chan bool

	// attemptTracker is used to track the information related to order-attempts.
	attemptTracker api.AttemptTracker

	// checkpointTracker is used to track the checkpoints for order-attempts.
	checkpointTracker api.CheckpointTracker

	// crypto is used to generate/verify signatures.
	crypto api.Crypto

	// sender is used to send network messages.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

// Run is used to start the processor for consensus messages in phalanx.
func (con *consensusRunner) Run() {
	for {
		select {
		case <-con.closeC:
			return
		case msg := <-con.consensusMessageC:
			if err := con.dispatchConsensusMessage(msg); err != nil {
				con.logger.Errorf("[%s] event error: %s", err)
			}
		}
	}
}

// Quit is used to stop the process.
func (con *consensusRunner) Quit() {
	select {
	case <-con.closeC:
	default:
		close(con.closeC)
	}
}

// ProcessConsensusMessage is used to process consensus messages.
func (con *consensusRunner) ProcessConsensusMessage(message *protos.ConsensusMessage) {
	con.consensusMessageC <- message
}

func (con *consensusRunner) dispatchConsensusMessage(message *protos.ConsensusMessage) error {
	if message == nil {
		return fmt.Errorf("nil message")
	}
	switch message.Type {
	case protos.MessageType_ORDER_ATTEMPT:
		return con.processOrderAttempt(message)
	case protos.MessageType_CHECKPOINT_REQUEST:
		return con.processCheckpointRequest(message)
	case protos.MessageType_CHECKPOINT_VOTE:
		return con.processCheckpointVote(message)
	default:
		return nil
	}
}

func (con *consensusRunner) processOrderAttempt(message *protos.ConsensusMessage) error {
	attempt := &protos.OrderAttempt{}
	if err := proto.Unmarshal(message.Payload, attempt); err != nil {
		return fmt.Errorf("unmarshal failed: %s", err)
	}

	if _, ok := con.aggMap[attempt.Digest]; ok {
		return nil
	}

	checkpoint := protos.NewCheckpoint(attempt)
	signature, err := con.crypto.PrivateSign(types.StringToBytes(attempt.Digest))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}
	if err = checkpoint.CombineQC(con.nodeID, signature); err != nil {
		return fmt.Errorf("combine QC failed: %s", err)
	}

	request := protos.NewCheckpointRequest(con.nodeID, attempt)
	cm, err := protos.PackCheckpointRequest(request)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	con.sender.BroadcastPCM(cm)
	con.aggMap[attempt.Digest] = checkpoint

	return nil
}

func (con *consensusRunner) processCheckpointRequest(message *protos.ConsensusMessage) error {
	request := &protos.CheckpointRequest{}
	if err := proto.Unmarshal(message.Payload, request); err != nil {
		return fmt.Errorf("unmarshal failed: %s", err)
	}

	vote := protos.NewCheckpointVote(con.nodeID, request)
	signature, err := con.crypto.PrivateSign(types.StringToBytes(request.OrderAttempt.Digest))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}
	vote.Cert = signature
	cm, err := protos.PackCheckpointVote(vote, request.Author)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	con.sender.BroadcastPCM(cm)
	return nil
}

func (con *consensusRunner) processCheckpointVote(message *protos.ConsensusMessage) error {
	vote := &protos.CheckpointVote{}
	if err := proto.Unmarshal(message.Payload, vote); err != nil {
		return fmt.Errorf("unmarshal failed: %s", err)
	}

	checkpoint, ok := con.aggMap[vote.Digest]
	if !ok {
		// there are 2 conditions that a pre-order waiting for agg-sig cannot be found: 1) we have never generated such
		// a pre-order message, 2) we have already generated an order message for it, which means it has been verified.
		return nil
	}

	// verify the signature in vote
	// here, we would like to check if the signature is valid.
	if err := con.crypto.PublicVerify(vote.Cert, types.StringToBytes(vote.Digest), vote.Author); err != nil {
		return fmt.Errorf("failed to aggregate: %s", err)
	}

	// record the certification in current vote
	if err := checkpoint.CombineQC(vote.Author, vote.Cert); err != nil {
		return fmt.Errorf("combine QC failed: %s", err)
	}

	// check the quorum size for proof-certs
	if checkpoint.IsValid(con.quorum) {
		con.logger.Debugf("[%d] found quorum votes, generate quorum order %s", con.nodeID, checkpoint.Format())
		delete(con.aggMap, vote.Digest)
		con.checkpointTracker.Record(checkpoint)
	}
	return nil
}
