package types

import "github.com/google/btree"

const (
	// OrderStatusPreOrder indicates the partial order is in pre-order status.
	OrderStatusPreOrder = iota

	// OrderStatusQuorumVerified indicates the partial order has been verified by quorum participants.
	OrderStatusQuorumVerified
)

// OrderEvent is used to process the verification events for the partial orders from each other.
type OrderEvent struct {
	// Status indicates the status of current partial order.
	Status int

	// Sequence indicates the assigned sequence number for current partial order.
	Sequence uint64

	// Digest indicates the identifier for current partial order.
	Digest string

	// Event contains the latest received message for current partial order.
	Event interface{}
}

func (event *OrderEvent) Less(item btree.Item) bool {
	return event.Sequence < (item.(*OrderEvent)).Sequence
}
