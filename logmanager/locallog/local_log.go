package locallog

import (
	"fmt"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type localLog struct {
	// author is the identifier for current node.
	author uint64

	// quorum is the legal size for current node.
	quorum int

	// sequence is the the target for local-log.
	sequence uint64

	// aggMap is used to generate aggregated-certificates.
	aggMap map[string]*protos.QuorumCert

	// logger is used to print logs.
	logger external.Logger
}

func NewLocalLog(n int, author uint64, logger external.Logger) *localLog {
	return &localLog{
		quorum:   types.CalculateQuorum(n),
		author:   author,
		sequence: uint64(0),
		aggMap:   make(map[string]*protos.QuorumCert),
		logger:   logger,
	}
}

// ProcessCommand is used to process command received from clients.
// We would like to assign a sequence number for such a command and generate a pre-order message.
func (local *localLog) ProcessCommand(command *protos.Command) (*protos.PreOrder, error) {
	local.sequence++

	pre := protos.NewPreOrder(local.author, local.sequence, command)
	payload, err := proto.Marshal(pre)
	if err != nil {
		local.logger.Errorf("Marshal Error: %v", err)
		return nil, err
	}
	pre.Digest = types.CalculatePayloadHash(payload, 0)

	// generate self-signature for current pre-order
	signature, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(local.author))
	if err != nil {
		return nil, fmt.Errorf("generate signature for pre-order failed: %s", err)
	}

	// init the order message in aggregate map and assign self signature
	local.aggMap[pre.Digest] = protos.NewQuorumCert(pre)
	local.aggMap[pre.Digest].ProofCerts.Certs[local.author] = signature

	local.logger.Infof("replica %d generated a pre-order, sequence %d, hash %s", local.author, pre.Sequence, pre.Digest)
	return pre, nil
}

// ProcessVote is used to process the vote message from others.
// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
func (local *localLog) ProcessVote(vote *protos.Vote) (*protos.QuorumCert, error) {
	local.logger.Debugf("replica %d received votes for %s from replica %d", local.author, vote.Digest, vote.Author)

	// check the existence of order message
	// here, we should make sure that there is a valid pre-order for us which we have ever assigned.
	order, ok := local.aggMap[vote.Digest]
	if !ok {
		// there are 2 conditions that a pre-order waiting for agg-sig cannot be found: 1) we have never generated such
		// a pre-order message, 2) we have already generated an order message for it, which means it has been verified.
		return nil, nil
	}

	// verify the signature in vote
	// here, we would like to check if the signature is valid.
	if err := crypto.PubVerify(vote.Certification, types.StringToBytes(vote.Digest), int(vote.Author)); err != nil {
		return nil, fmt.Errorf("failed to aggregate: %s", err)
	}

	// record the certification in current vote
	order.ProofCerts.Certs[vote.Author] = vote.Certification

	// check the quorum size for proof-certs
	if len(order.ProofCerts.Certs) == local.quorum {
		local.logger.Debugf("replica %d find quorum votes for pre-log sequence %d hash %s, generate quorum order",
			local.author, order.PreOrder.Sequence, order.PreOrder.Digest)
		delete(local.aggMap, vote.Digest)
		return order, nil
	}

	local.logger.Debugf("replica %d aggregated vote for %s, has %d, need %d", local.author, vote.Digest, len(order.ProofCerts.Certs), local.quorum)
	return nil, nil
}
