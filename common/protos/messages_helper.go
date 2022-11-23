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

func PackOrderAttempt(attempt *OrderAttempt) (*ConsensusMessage, error) {
	payload, err := proto.Marshal(attempt)
	if err != nil {
		return nil, err
	}
	return NewConsensusMessage(MessageType_ORDER_ATTEMPT, attempt.NodeID, 0, payload), nil
}

//=============================== Command ===============================================

func (m *Command) Less(item btree.Item) bool {
	return m.Sequence < (item.(*Command)).Sequence
}

func (m *Command) Format() string {
	return fmt.Sprintf("[Command: client %d, sequence %d, digest %s, front-runner: %v]", m.Author, m.Sequence, m.Digest, m.FrontRunner)
}

func (m *Command) FormatSnappingUp() string {
	return fmt.Sprintf("[Command: buyer %d, itemNo %d, digest %s]", m.Author, m.Sequence, m.Digest)
}

//=============================== Pre-Order ===============================================

func (m *PreOrder) Format() string {
	return fmt.Sprintf("[PreOrder: author %d, sequence %d, digest %s, command list %v, timestamp list %v]", m.Author, m.Sequence, m.Digest, m.CommandList, m.TimestampList)
}

//=============================== Vote ====================================================

func (m *Vote) Format() string {
	return fmt.Sprintf("[Vote: author %d, vote-digest %s]", m.Author, m.Digest)
}

//=============================== Partial Order ===============================================

func (m *PartialOrder) Author() uint64 {
	return m.PreOrder.Author
}

func (m *PartialOrder) PreOrderDigest() string {
	return m.PreOrder.Digest
}

func (m *PartialOrder) CommandList() []string {
	return m.PreOrder.CommandList
}

func (m *PartialOrder) Sequence() uint64 {
	return m.PreOrder.Sequence
}

func (m *PartialOrder) TimestampList() []int64 {
	return m.PreOrder.TimestampList
}

func (m *PartialOrder) SetOrderedTime() {
	m.OrderedTime = time.Now().UnixNano()
}

func (m *PartialOrder) Format() string {
	return fmt.Sprintf("[PartialOrder: author %d, sequence %d, command list %v, timestamp list %v]", m.Author(), m.Sequence(), m.CommandList(), m.TimestampList())
}

func (m *PartialOrder) ParentDigest() string {
	return m.PreOrder.ParentDigest
}

//=================================== Partial Order Batch =========================================

func (m *PartialOrderBatch) Format() string {
	return fmt.Sprintf("[PartialBatch: author %d, proposed nos %v]", m.Author, m.SeqList)
}

func (m *OrderAttempt) Format() string {
	return fmt.Sprintf("[OrderAttempt: nodeID %d, seqNo %d, digest %s, parentDigest %s]", m.NodeID, m.SeqNo, m.Digest, m.ParentDigest)
}

func (m *OrderAttemptContent) Format() string {
	return fmt.Sprintf("[OrderAttemptContent: commandList %v, timestampList %v]", m.CommandList, m.TimestampList)
}

//=================================== Generate Messages ============================================

func NewQuorumCert() *QuorumCert {
	return &QuorumCert{Certs: make(map[uint64]*Certification)}
}

func NewPartialOrder(pre *PreOrder) *PartialOrder {
	return &PartialOrder{PreOrder: pre, QC: NewQuorumCert()}
}

func NewNopPartialOrder() *PartialOrder {
	return &PartialOrder{PreOrder: NewNopPreOrder(), QC: NewQuorumCert()}
}

func NewPreOrder(author uint64, sequence uint64, commandList []string, timestampList []int64, previous *PreOrder) *PreOrder {
	if previous == nil {
		previous = &PreOrder{Digest: "GENESIS PRE ORDER"}
	}
	return &PreOrder{Author: author, Sequence: sequence, CommandList: commandList, TimestampList: timestampList, ParentDigest: previous.Digest}
}

func NewNopPreOrder() *PreOrder {
	return &PreOrder{Digest: "Nop Pre Order"}
}

func NewPartialOrderBatch(author uint64, count int) *PartialOrderBatch {
	return &PartialOrderBatch{Author: author, HighOrders: make([]*PartialOrder, count), SeqList: make([]uint64, count)}
}

func NewOrderAttemptContent(commandList []string, timestampList []int64) *OrderAttemptContent {
	return &OrderAttemptContent{CommandList: commandList, TimestampList: timestampList}
}

func NewOrderAttempt(nodeID uint64, seqNo uint64, previous *OrderAttempt, contentDigest string, content *OrderAttemptContent) *OrderAttempt {
	if previous == nil {
		previous = &OrderAttempt{Digest: "GENESIS PRE ORDER"}
	}
	return &OrderAttempt{NodeID: nodeID, SeqNo: seqNo, ParentDigest: previous.Digest, ContentDigest: contentDigest, Content: content}
}
