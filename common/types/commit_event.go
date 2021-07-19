package types

import (
	"fmt"
	"github.com/Grivn/phalanx/common/protos"

	"github.com/google/btree"
)

type CommitEvent struct {
	Sequence  uint64
	PBatch    *protos.PartialOrderBatch
	IsValid   bool
}

func (ce *CommitEvent) Less(item btree.Item) bool {
	return ce.Sequence < (item.(*CommitEvent)).Sequence
}

func (ce *CommitEvent) Format() string {
	return fmt.Sprintf("[CommitEvent: sequence %d, valid %v]", ce.Sequence, ce.IsValid)
}
