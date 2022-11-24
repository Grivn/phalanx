package metapool

import (
	"fmt"
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type consensusRunner struct {
	nodeID uint64

	n int

	quorum int

	// aggMap is used to make aggregation quorum certificate for checkpoints.
	aggMap map[string]*protos.Checkpoint

	attemptC chan *protos.OrderAttempt

	requestC chan *protos.CheckpointRequest

	voteC chan *protos.CheckpointVote

	eventC chan interface{}

	closeC chan bool

	aTracker api.AttemptTracker

	crypto api.Crypto
	sender external.NetworkService
	logger external.Logger
}

func (con *consensusRunner) run() {
	for {
		select {
		case <-con.closeC:
			return
		case ev := <-con.eventC:
			if err := con.dispatchEvents(ev); err != nil {
				con.logger.Errorf("[%s] event error: %s", err)
			}
		}
	}
}

func (con *consensusRunner) quit() {
	select {
	case <-con.closeC:
	default:
		close(con.closeC)
	}
}

func (con *consensusRunner) dispatchEvents(ev interface{}) error {
	switch event := ev.(type) {
	case *protos.OrderAttempt:
		return con.requestCheckpoint(event)
	case *protos.CheckpointRequest:
		return con.processCheckpointRequest(event)
	case *protos.CheckpointVote:
		return con.processCheckpointVote(event)
	default:
		return nil
	}
}

func (con *consensusRunner) requestCheckpoint(attempt *protos.OrderAttempt) error {
	if _, ok := con.aggMap[attempt.Digest]; ok {
		return nil
	}

	checkpoint := protos.NewCheckpoint(attempt)
	signature, err := con.crypto.PrivateSign(types.StringToBytes(attempt.Digest))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}
	checkpoint.Certs()[con.nodeID] = signature

	request := protos.NewCheckpointRequest(con.nodeID, attempt)
	cm, err := protos.PackCheckpointRequest(request)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	con.sender.BroadcastPCM(cm)
	con.aggMap[attempt.Digest] = checkpoint

	return nil
}

func (con *consensusRunner) processCheckpointRequest(request *protos.CheckpointRequest) error {

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

func (con *consensusRunner) processCheckpointVote(vote *protos.CheckpointVote) error {
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
	checkpoint.Certs()[vote.Author] = vote.Cert

	// check the quorum size for proof-certs
	if len(checkpoint.Certs()) == con.quorum {
		con.logger.Debugf("[%d] found quorum votes, generate quorum order %s", con.nodeID, checkpoint.Format())
		delete(con.aggMap, vote.Digest)
		con.aTracker.Checkpoint(checkpoint)
	}
	return nil
}
