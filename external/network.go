package external

import "github.com/Grivn/phalanx/common/protos"

// NetworkService has provided the service for network communication.
type NetworkService interface {
	// BroadcastPCM is used to send the message to every node in current cluster.
	// in this implement of phalanx, we prefer a PhalanxBroadcast in which we could send message to ourself.
	BroadcastPCM(message *protos.ConsensusMessage)

	// UnicastPCM is used to send the message to the target node.
	UnicastPCM(message *protos.ConsensusMessage)
}
