package api

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type MetaPoolV2 interface {
	Runner
	Sequencer

	// GenerateProposal generates proposal for total consensus processor.
	GenerateProposal() (*protos.PartialOrderBatch, error)

	// VerifyProposal verifies the proposal consented by consensus processor
	// and generate query stream for partial orders if essential.
	VerifyProposal(batch *protos.PartialOrderBatch) (types.QueryStream, error)
}

type Sequencer interface {
	// GenerateOptimisticOrder is used to receive commands from clients
	// and assign incremental sequence number on current node.
	GenerateOptimisticOrder(command *protos.Command)
}
