package protos

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/btree"
)

//=============================== Consensus Message ===============================================

func NewConsensusMessage(typ MessageType, from, to uint64, payload []byte) *ConsensusMessage {
	return &ConsensusMessage{Type: typ, From: from, To: to, Payload: payload}
}

func PackPreOrder(pre *PreOrder) (*ConsensusMessage, error) {
	payload, err := proto.Marshal(pre)
	if err != nil {
		return nil, err
	}
	return NewConsensusMessage(MessageType_PRE_ORDER, pre.Author, 0, payload), nil
}

func PackVote(vote *Vote, to uint64) (*ConsensusMessage, error) {
	payload, err := proto.Marshal(vote)
	if err != nil {
		return nil, err
	}
	return NewConsensusMessage(MessageType_VOTE, vote.Author, to, payload), nil
}

func PackPartialOrder(qc *PartialOrder) (*ConsensusMessage, error) {
	payload, err := proto.Marshal(qc)
	if err != nil {
		return nil, err
	}
	return NewConsensusMessage(MessageType_QUORUM_CERT, qc.Author(), 0, payload), nil
}
//=============================== Partial Order ===============================================

func (m *PartialOrder) Less(item btree.Item) bool {
	// for b-tree initiation
	return m.PreOrder.Sequence < (item.(*PartialOrder)).PreOrder.Sequence
}

func (m *PartialOrder) Author() uint64 {
	return m.PreOrder.Author
}

func (m *PartialOrder) Digest() string {
	return m.PreOrder.Digest
}

func (m *PartialOrder) CommandDigest() string {
	return m.PreOrder.BatchDigest
}

func (m *PartialOrder) Sequence() uint64 {
	return m.PreOrder.Sequence
}

func (m *PartialOrder) Timestamp() int64 {
	return m.PreOrder.Timestamp
}

//=============================== Partial Order Batch ===============================================

func (m *PartialOrderBatch) Append(pOrder *PartialOrder) {
	// append:
	// we have found a partial order which could be proposed in next phase, append into Partials slice.
	m.Partials = append(m.Partials, pOrder)
}

//=================================== Generate Messages ============================================

func NewQuorumCert() *QuorumCert {
	return &QuorumCert{Certs: make(map[uint64]*Certification)}
}


func NewPartialOrder(pre *PreOrder) *PartialOrder {
	return &PartialOrder{PreOrder: pre, QC: NewQuorumCert()}
}

func NewPreOrder(author uint64, sequence uint64, command *Command) *PreOrder {
	return &PreOrder{Author: author, Sequence: sequence, BatchDigest: command.Digest, Timestamp: time.Now().UnixNano()}
}

func NewPartialOrderBatch(author uint64) *PartialOrderBatch {
	return &PartialOrderBatch{Author: author, Commands: make(map[string]*Command)}
}
