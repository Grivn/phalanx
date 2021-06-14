package phalanx

import "github.com/Grivn/phalanx/common/protos"

type Provider interface {
	Communicate
	Generator
	Validator
}

type Communicate interface {
	ProcessCommand(command *protos.Command)
	ProcessConsensusMessage(message *protos.ConsensusMessage)
}

type Generator interface {
	MakePayload() ([]byte, error)
}

type Validator interface {
	VerifyPayload(payload []byte) error
	StablePayload(payload []byte) error
	CommitPayload(payload []byte) error
}
