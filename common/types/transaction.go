package types

import (
	"math/rand"
	"time"

	"github.com/Grivn/phalanx/common/protos"

	"github.com/gogo/protobuf/proto"
)

//=================================== Command Generator =======================================

// GenerateCommand generates command with given transaction list.
func GenerateCommand(author uint64, seqNo uint64, txs []*protos.Transaction) *protos.Command {
	var hashList []string
	for _, tx := range txs {
		hashList = append(hashList, tx.Hash)
	}
	command := &protos.Command{
		Author:   author,
		Sequence: seqNo,
		HashList: hashList,
	}
	payload, err := proto.Marshal(command)
	if err != nil {
		return nil
	}
	command.Digest = CalculatePayloadHash(payload, 0)
	command.Content = txs
	return command
}

func GenerateRandCommand(author uint64, seqNo uint64, count, size int) *protos.Command {
	tList := make([]*protos.Transaction, count)
	hList := make([]string, count)

	for i:=0; i<count; i++ {
		tx := GenerateRandTransaction(size)

		tList[i] = tx
		hList[i] = tx.Hash
	}

	command := &protos.Command{Author: author, Sequence: seqNo, HashList: hList}
	payload, err := proto.Marshal(command)
	if err != nil {
		panic(err)
	}
	command.Digest = CalculatePayloadHash(payload, 0)
	command.Content = tList

	return command
}

//==================================== Transaction Generator ====================================

func GenerateRandTransaction(size int) *protos.Transaction {
	payload := make([]byte, size)
	rand.Read(payload)
	return GenerateTransaction(payload)
}

func GenerateTransaction(payload []byte) *protos.Transaction {
	return &protos.Transaction{
		Hash:      CalculatePayloadHash(payload, time.Now().UnixNano()),
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
	}
}
