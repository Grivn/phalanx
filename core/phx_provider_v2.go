package phalanx

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

// ProviderV2 is the phalanx service provider for all kinds of consensus algorithm, such as PBFT or HS.
type ProviderV2 interface {
	Run()

	Quit()

	// ReceiveTransaction is used to process transaction we have received.
	ReceiveTransaction(tx *protos.Transaction)

	// ReceiveCommand is used to process the commands from clients.
	ReceiveCommand(command *protos.Command)

	// ReceiveOrderAttempt is used to process the order-attempts from sequencers.
	ReceiveOrderAttempt(attempt *protos.OrderAttempt)

	// ReceiveConsensusMessage is used process the consensus messages from phalanx replica.
	ReceiveConsensusMessage(message *protos.ConsensusMessage)

	// MakeProposal is used to generate phalanx proposal for consensus.
	MakeProposal() (*protos.Proposal, error)

	// CommitProposal is used to commit the phalanx proposal which has been verified with consensus.
	CommitProposal(pBatch *protos.Proposal) error

	// QueryMetrics returns the metrics info of phalanx.
	QueryMetrics() types.MetricsInfo
}
