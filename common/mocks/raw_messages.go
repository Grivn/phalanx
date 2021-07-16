package mocks

import (
	"math/rand"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"

	"github.com/gogo/protobuf/proto"
)

func NewTransaction() *protos.PTransaction {
	payload := make([]byte, 4)
	rand.Read(payload)
	return &protos.PTransaction{Hash: types.CalculatePayloadHash(payload, 0), Payload: payload}
}

func NewCommand() *protos.PCommand {
	tx := NewTransaction()
	txList := []*protos.PTransaction{tx}
	hashList := []string{tx.Hash}
	command := &protos.PCommand{Content: txList, HashList: hashList}

	payload, _ := proto.Marshal(command)
	command.Digest = types.CalculatePayloadHash(payload, 0)

	return command
}
