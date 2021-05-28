package protos

import (
	"fmt"
	"github.com/google/btree"
)

func (m *Proposal) Format() string {
	return fmt.Sprintf("[Proposal, Author %d, Sequence %d, Batch %s]", m.Author, m.Sequence, m.TxBatch.Digest)
}
