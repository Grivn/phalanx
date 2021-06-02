package reliablelog

import "github.com/Grivn/phalanx/common/protos"

type SubInstance interface {
	// SubID is used to get the identifier for current sub-instance.
	SubID() uint64

	// ProcessPreOrder is used to process pre-order messages.
	// We should make sure that we have never received a pre-order/order message
	// whose sequence number is the same as it yet, and we would like to generate a
	// vote message for it if it's legal for us.
	ProcessPreOrder(pre *protos.PreOrder) error

	// ProcessOrder is used to process order messages.
	// A valid order message, which has a series of valid signature which has reached quorum size,
	// could advance the sequence counter. We should record the advanced counter and put the info of
	// order message into the sequential-pool.
	ProcessOrder(order *protos.Order) error
}
