package external

import "github.com/Grivn/phalanx/common/protos"

type NetworkService interface {
	Broadcast(message *protos.ConsensusMessage)
	Unicast(message *protos.ConsensusMessage)
}
