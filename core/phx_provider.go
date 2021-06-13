package phalanx

import "github.com/Grivn/phalanx/common/protos"

type Provider interface {
	ProcessCommand(command *protos.Command)
	ProcessConsensusMessage(message *protos.ConsensusMessage)
	MakePayload() ([]byte, error)
	VerifyPayload(payload []byte) error
	StablePayload(payload []byte) error
	CommitPayload(payload []byte) error
}
