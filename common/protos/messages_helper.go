package protos

import (
	"fmt"
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

//=============================== Command ===============================================

func (m *Command) Less(item btree.Item) bool {
	return m.Sequence < (item.(*Command)).Sequence
}

func (m *Command) Format() string {
	return fmt.Sprintf("[Command: client %d, sequence %d, digest %s]", m.Author, m.Sequence, m.Digest)
}

//=============================== Pre-Order ===============================================

func (m *PreOrder) Format() string {
	return fmt.Sprintf("[PreOrder: author %d, sequence %d, digest %s, command %s]", m.Author, m.Sequence, m.Digest, m.CommandDigest)
}

//=============================== Vote ====================================================

func (m *Vote) Format() string {
	return fmt.Sprintf("[Vote: author %d, vote-digest %s]", m.Author, m.Digest)
}

//=============================== Partial Order ===============================================

func (m *PartialOrder) Less(item btree.Item) bool {
	// for b-tree initiation
	return m.PreOrder.Sequence < (item.(*PartialOrder)).PreOrder.Sequence
}

func (m *PartialOrder) Author() uint64 {
	return m.PreOrder.Author
}

func (m *PartialOrder) PreOrderDigest() string {
	return m.PreOrder.Digest
}

func (m *PartialOrder) CommandDigest() string {
	return m.PreOrder.CommandDigest
}

func (m *PartialOrder) Sequence() uint64 {
	return m.PreOrder.Sequence
}

func (m *PartialOrder) Timestamp() int64 {
	return m.PreOrder.Timestamp
}

func (m *PartialOrder) Format() string {
	return fmt.Sprintf("[PartialOrder: author %d, sequence %d, command %s]", m.Author(), m.Sequence(), m.CommandDigest())
}

func (m *PartialOrder) ParentDigest() string {
	return m.PreOrder.ParentDigest
}

//=============================== Partial Order Batch ===============================================

func (m *PartialOrderBatch) Append(pOrder *PartialOrder) {
	// append:
	// we have found a partial order which could be proposed in next phase, append into Partials slice.
	//m.Partials = append(m.Partials, pOrder)
	//m.ProposedNos[pOrder.Author()] = maxUint64(m.ProposedNos[pOrder.Author()], pOrder.Sequence())
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

//=================================== Generate Messages ============================================

func NewQuorumCert() *QuorumCert {
	return &QuorumCert{Certs: make(map[uint64]*Certification)}
}

func NewPartialOrder(pre *PreOrder) *PartialOrder {
	return &PartialOrder{PreOrder: pre, QC: NewQuorumCert()}
}

func NewPreOrder(author uint64, sequence uint64, command *Command, previous *PreOrder) *PreOrder {
	if previous == nil {
		previous = &PreOrder{Digest: "GENESIS PRE ORDER"}
	}
	return &PreOrder{Author: author, Sequence: sequence, CommandDigest: command.Digest, Timestamp: time.Now().UnixNano(), ParentDigest: previous.Digest}
}

func NewPartialOrderBatch(author uint64) *PartialOrderBatch {
	return &PartialOrderBatch{Author: author, Partials: make(map[uint64]*PartialOrder), Commands: make(map[string]*Command), ProposedNos: make(map[uint64]uint64)}
}
