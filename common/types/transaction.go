package types

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"math/rand"

	"github.com/Grivn/phalanx/common/protos"

	"github.com/gogo/protobuf/proto"
)

func GenerateRandCommand(count, size int) *protos.Command {
	tList := make([]*protos.Transaction, count)
	hList := make([]string, count)

	for i:=0; i<count; i++ {
		tx := GenerateRandTransaction(size)

		tList[i] = tx
		hList[i] = tx.Hash
	}

	command := &protos.Command{Content: tList, HashList: hList}
	payload, err := proto.Marshal(command)
	if err != nil {
		panic(err)
	}
	command.Digest = CalculatePayloadHash(payload, 0)

	return command
}

func GenerateRandTransaction(size int) *protos.Transaction {
	payload := make([]byte, size)
	rand.Read(payload)
	return GenerateTransaction(payload)
}

func GenerateTransaction(payload []byte) *protos.Transaction {
	return &protos.Transaction{
		Hash:    CalculatePayloadHash(payload, 0),
		Payload: payload,
	}
}

// GetHash returns the TransactionHash
func GetHash(tx *protos.Transaction) string {
	if tx.Hash == "" {
		tx.Hash = CalculatePayloadHash(tx.Payload, 0)
	}
	return tx.Hash
}

func CalculateListHash(list []string, timestamp int64) string {
	h := md5.New()
	for _, hash := range list {
		_, _ = h.Write([]byte(hash))
	}

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}
	return BytesToString(h.Sum(nil))
}

func CalculateMD5Hash(payload []byte, timestamp int64) []byte {
	h := md5.New()
	_, _ = h.Write(payload)

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}
	return h.Sum(nil)
}

func CalculatePayloadHash(payload []byte, timestamp int64) string {
	h := md5.New()
	_, _ = h.Write(payload)

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}
	return BytesToString(h.Sum(nil))
}

func BytesToString(b []byte) string {
	return hex.EncodeToString(b)
}

func StringToBytes(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

func Uint64MapToList(m map[uint64]bool) []uint64 {
	var list []uint64
	for key := range m {
		list = append(list, key)
	}
	return list
}

func Uint64ToBytes(num uint64) []byte {
	buffer := bytes.NewBuffer([]byte{})

	err := binary.Write(buffer, binary.BigEndian, &num)
	if err != nil {
		panic("error convert bytes to int")
	}

	return buffer.Bytes()
}

func BytesToUint64(bys []byte) uint64 {
	buffer := bytes.NewBuffer([]byte{})

	var num uint64
	err := binary.Read(buffer, binary.BigEndian, &num)
	if err != nil {
		panic("error convert bytes to int")
	}

	return num
}
