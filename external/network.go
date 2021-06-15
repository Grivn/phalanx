package external

import "github.com/Grivn/phalanx/common/protos"

// NetworkService has provided the service for network communication.
type NetworkService interface {
	// Broadcast is used to send the message to every node in current cluster.
	// in this implement of phalanx, we prefer a Broadcast in which we could send message to ourself.
	Broadcast(message *protos.ConsensusMessage)

	// Unicast is used to send the message to the target node.
	Unicast(message *protos.ConsensusMessage)
}
