package conengine

import (
	"fmt"
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/gogo/protobuf/proto"
)

type consensusEngine struct {
	// nodeID is the identifier of current node.
	nodeID uint64

	// n is the total number of the participants.
	n int

	// quorum is the threshold of stable consensus.
	quorum int

	// aggregateMap is used to create aggregated-QC for checkpoints.
	aggregateMap map[string]*protos.Checkpoint

	// eventC is used to receive event messages from local modules.
	eventC chan types.LocalEvent

	// consensusMessageC is used to relay the consensus message into consensus runner.
	consensusMessageC chan *protos.ConsensusMessage

	// closeC is used to stop the processor of consensus runner.
	closeC chan bool

	// checkpointTracker is used to track the checkpoints for order-attempts.
	checkpointTracker api.CheckpointTracker

	// crypto is used to generate/verify signatures.
	crypto api.Crypto

	// sender is used to send network messages.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func NewConsensusEngine(conf Config) api.ConsensusEngine {
	return &consensusEngine{
		nodeID:            conf.NodeID,
		n:                 conf.N,
		quorum:            types.CalculateQuorum(conf.N),
		aggregateMap:      make(map[string]*protos.Checkpoint),
		eventC:            make(chan types.LocalEvent),
		consensusMessageC: make(chan *protos.ConsensusMessage),
		closeC:            make(chan bool),
		sender:            conf.External,
		logger:            conf.External,
	}
}

// Run is used to start the processor for consensus messages in phalanx.
func (con *consensusEngine) Run() {
	for {
		select {
		case <-con.closeC:
			return
		case message := <-con.consensusMessageC:
			if err := con.dispatchConsensusMessage(message); err != nil {
				con.logger.Errorf("[%d] process consensus message error: %s", con.nodeID, err)
			}
		case event := <-con.eventC:
			if err := con.dispatchLocalEvent(event); err != nil {
				con.logger.Errorf("[%d] process local event error: %s", con.nodeID, err)
			}
		}
	}
}

// Quit is used to stop the process.
func (con *consensusEngine) Quit() {
	select {
	case <-con.closeC:
	default:
		close(con.closeC)
	}
}

func (con *consensusEngine) ProcessConsensusMessage(message *protos.ConsensusMessage) {
	con.consensusMessageC <- message
}

func (con *consensusEngine) ProcessLocalEvent(event types.LocalEvent) {
	con.eventC <- event
}

func (con *consensusEngine) dispatchConsensusMessage(message *protos.ConsensusMessage) error {
	if message == nil {
		return fmt.Errorf("nil message")
	}
	switch message.Type {
	case protos.MessageType_CHECKPOINT_REQUEST:
		request := &protos.CheckpointRequest{}
		if err := proto.Unmarshal(message.Payload, request); err != nil {
			return fmt.Errorf("unmarshal failed: %s", err)
		}
		return con.processCheckpointRequest(request)
	case protos.MessageType_CHECKPOINT_VOTE:
		vote := &protos.CheckpointVote{}
		if err := proto.Unmarshal(message.Payload, vote); err != nil {
			return fmt.Errorf("unmarshal failed: %s", err)
		}
		return con.processCheckpointVote(vote)
	default:
		return nil
	}
}

func (con *consensusEngine) dispatchLocalEvent(event types.LocalEvent) error {
	switch event.Type {
	case types.LocalEventOrderAttempt:
		attempt, ok := event.Event.(*protos.OrderAttempt)
		if !ok {
			return fmt.Errorf("parse order attempt failed")
		}
		if err := con.processOrderAttempt(attempt); err != nil {
			return fmt.Errorf("process order-attempt failed: %s", err)
		}
		return nil
	default:
		return fmt.Errorf("invalid event type %d", event.Type)
	}
}

func (con *consensusEngine) processOrderAttempt(attempt *protos.OrderAttempt) error {
	if attempt == nil {
		return fmt.Errorf("nil attempt")
	}

	if _, ok := con.aggregateMap[attempt.Digest]; ok {
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
	con.aggregateMap[attempt.Digest] = checkpoint

	return nil
}

func (con *consensusEngine) processCheckpointRequest(request *protos.CheckpointRequest) error {
	if request == nil {
		return fmt.Errorf("nil request")
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

func (con *consensusEngine) processCheckpointVote(vote *protos.CheckpointVote) error {
	if vote == nil {
		return fmt.Errorf("nil vote")
	}

	checkpoint, ok := con.aggregateMap[vote.Digest]
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
		delete(con.aggregateMap, vote.Digest)
		con.checkpointTracker.Record(checkpoint)
	}
	return nil
}
