package mocks

import (
	"math/rand"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"

	"github.com/gogo/protobuf/proto"
)

func NewTransaction() *protos.Transaction {
	payload := make([]byte, 4)
	rand.Read(payload)
	return &protos.Transaction{Hash: types.CalculatePayloadHash(payload, 0), Payload: payload}
}

func NewCommand() *protos.Command {
	tx := NewTransaction()
	txList := []*protos.Transaction{tx}
	hashList := []string{tx.Hash}
	command := &protos.Command{Content: txList, HashList: hashList}

	payload, _ := proto.Marshal(command)
	command.Digest = types.CalculatePayloadHash(payload, 0)

	return command
}
