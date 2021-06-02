package reliablelog

import (
	"fmt"
	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/types"
	"github.com/gogo/protobuf/proto"
	"time"

	commonProto "github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type orderMgr struct {
	quorum int

	author   uint64
	sequence uint64
	aggMap   map[string]*commonProto.Order

	logger external.Logger
}

func newOrderMgr(author uint64, logger external.Logger) *orderMgr {
	return &orderMgr{
		author:   author,
		sequence: uint64(0),
		logger:   logger,
	}
}

// generatePreOrder is used to generate pre-order for one command
func (om *orderMgr) generatePreOrder(command *commonProto.Command) (*commonProto.PreOrder, error) {
	om.sequence++

	pre := &commonProto.PreOrder{
		Author:      om.author,
		Sequence:    om.sequence,
		BatchDigest: command.Digest,
		Timestamp:   time.Now().UnixNano(),
	}

	payload, err := proto.Marshal(pre)
	if err != nil {
		om.logger.Errorf("Marshal Error: %v", err)
		return nil, err
	}
	pre.Digest = types.CalculatePayloadHash(payload, 0)

	// init the order message in aggregate map
	om.aggMap[pre.Digest] = &commonProto.Order{
		PreOrder: pre,
		ProofCerts:  make(map[uint64]*commonProto.Certification),
	}

	om.logger.Infof("replica %d generated a pre-log, sequence %d, hash %s", om.author, pre.Sequence, pre.Digest)
	return pre, nil
}

func (om *orderMgr) processVote(vote *commonProto.Vote) (*commonProto.Order, error) {
	om.logger.Debugf("replica %d received votes for %s from replica %d", om.author, vote.Digest, vote.Author)

	// check the existence of order message
	order, ok := om.aggMap[vote.Digest]
	if !ok {
		return nil, nil
	}

	// verify the signature in vote
	if err := crypto.PubVerify(vote.Certification, types.StringToBytes(vote.Digest), int(vote.Author)); err != nil {
		return nil, fmt.Errorf("failed to aggregate: %s", err)
	}

	// record the certification in current vote
	order.ProofCerts[vote.Author] = vote.Certification

	if len(order.ProofCerts) == om.quorum {
		om.logger.Debugf("replica %d find quorum votes for pre-log sequence %d hash %s, generate quorum order",
			om.author, order.PreOrder.Sequence, order.PreOrder.Digest)
		delete(om.aggMap, vote.Digest)
		return order, nil
	}

	om.logger.Debugf("replica %d aggregated vote for %s, has %d, need %d", om.author, vote.Digest, len(order.ProofCerts), om.quorum)
	return nil, nil
}
