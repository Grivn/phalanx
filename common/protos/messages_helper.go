package protos

import (
	"fmt"
	"time"

	"github.com/google/btree"
)

func (m *Proposal) Format() string {
	return fmt.Sprintf("[Proposal, Author %d, Sequence %d, Batch %s]", m.Author, m.Sequence, m.TxBatch.Digest)
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
