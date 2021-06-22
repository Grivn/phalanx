package external

import "github.com/Grivn/phalanx/common/protos"

// NetworkService has provided the service for network communication.
type NetworkService interface {
	// PhalanxBroadcast is used to send the message to every node in current cluster.
	// in this implement of phalanx, we prefer a PhalanxBroadcast in which we could send message to ourself.
	PhalanxBroadcast(message *protos.ConsensusMessage)

	// PhalanxUnicast is used to send the message to the target node.
	PhalanxUnicast(message *protos.ConsensusMessage)
}
