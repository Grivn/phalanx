package types

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"

	"github.com/Grivn/phalanx/common/protos"

	"github.com/gogo/protobuf/proto"
)

//==================================== Validator =============================================

const (
	COUNT = 512
	// Supported signing algorithms
	BLS_BLS12381    = "BLS_BLS12381"
	ECDSA_P256      = "ECDSA_P256"
	ECDSA_SECp256k1 = "ECDSA_SECp256k1"
)

type Signature [][]byte

//==================================== Hasher =============================================

// Identifier is the id for payloads
type Identifier [32]byte

// Hash is the hash algorithms output types
type Hash []byte

// Equal checks if a hash is equal to a given hash
func (h Hash) Equal(input Hash) bool {
	return bytes.Equal(h, input)
}

// Hex returns the hex string representation of the hash.
func (h Hash) Hex() string {
	return hex.EncodeToString(h)
}

//================================== Hash Management ===========================================

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

func CalculateBatchHash(pBatch *protos.PartialOrderBatch) string {
	payload, _ := proto.Marshal(pBatch)
	return CalculatePayloadHash(payload, 0)
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
