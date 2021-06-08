package types

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"github.com/Grivn/phalanx/common/protos"
)

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
		h.Write(b)
	}
	return BytesToString(h.Sum(nil))
}

func CalculateMD5Hash(payload []byte, timestamp int64) []byte {
	h := md5.New()
	h.Write(payload)

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		h.Write(b)
	}
	return h.Sum(nil)
}

func CalculatePayloadHash(payload []byte, timestamp int64) string {
	h := md5.New()
	h.Write(payload)

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		h.Write(b)
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
