package protos

import "fmt"

func (m *Proposal) Format() string {
	return fmt.Sprintf("[Proposal, Author %d, Sequence %d, Batch %s]", m.Author, m.Sequence, m.TxBatch.Digest)
}
