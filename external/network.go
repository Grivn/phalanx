package external

import "github.com/Grivn/phalanx/common/protos"

type Network interface {
	Broadcast(message *protos.ConsensusMessage)
	Unicast(message *protos.ConsensusMessage)
}
