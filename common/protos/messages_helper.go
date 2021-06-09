package protos

import (
	"github.com/gogo/protobuf/proto"
	"time"

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

func PackQC(qc *QuorumCert) (*ConsensusMessage, error) {
	payload, err := proto.Marshal(qc)
	if err != nil {
		return nil, err
	}
	return NewConsensusMessage(MessageType_QUORUM_CERT, qc.Author(), 0, payload), nil
}
//=============================== Quorum Cert ===============================================

func (m *QuorumCert) Less(item btree.Item) bool {
	return m.PreOrder.Sequence < (item.(*QuorumCert)).PreOrder.Sequence
}

func (m *QuorumCert) Author() uint64 {
	return m.PreOrder.Author
}

func (m *QuorumCert) Digest() string {
	return m.PreOrder.Digest
}

func (m *QuorumCert) Sequence() uint64 {
	return m.PreOrder.Sequence
}

func (m *QuorumCert) Timestamp() int64 {
	return m.PreOrder.Timestamp
}

//=================================== Generate Messages ============================================

func NewProofCerts() *ProofCerts {
	return &ProofCerts{Certs: make(map[uint64]*Certification)}
}

func NewQuorumCert(pre *PreOrder) *QuorumCert {
	return &QuorumCert{PreOrder: pre, ProofCerts: NewProofCerts()}
}

func NewPreOrder(author uint64, sequence uint64, command *Command) *PreOrder {
	return &PreOrder{Author: author, Sequence: sequence, BatchDigest: command.Digest, Timestamp: time.Now().UnixNano()}
}

func NewQCFilter() *QCFilter {
	return &QCFilter{QCs: nil}
}

func NewQCBatch() *QCBatch {
	return &QCBatch{Commands: make(map[string]*Command)}
}
