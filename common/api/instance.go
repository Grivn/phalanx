package api

import "github.com/Grivn/phalanx/common/protos"

//==================================== instance for meta pool =============================================

// ClientInstance is used to process commands info generated by specific client.
type ClientInstance interface {
	// Commit is used to notify the instance the committed command generated by current client.
	Commit(seqNo uint64) int

	// Append is used to notify the latest received command from current client.
	Append(command *protos.Command) int
}

// ReplicaInstance is used to process partial orders generated by each participant.
type ReplicaInstance interface {
	// GetHighOrder returns the highest partial order we have verified for current replica.
	GetHighOrder() *protos.PartialOrder

	// ReceivePreOrder is used to process the pre-order message from current replica.
	ReceivePreOrder(pre *protos.PreOrder) error

	// ReceivePartial is used to process the partial order message from current replica.
	ReceivePartial(pOrder *protos.PartialOrder) error
}

type SequencerInstance interface {
	// GetHighestAttempt gets the latest legal order-attempt received from current sequencer.
	GetHighestAttempt() *protos.OrderAttempt

	// Append is used to notify the latest received order-attempt from current sequencer and try to verify it.
	Append(attempt *protos.OrderAttempt)
}
