package types

import (
	"bytes"
	"crypto/sha256" // There is vulnerable problem for "crypto/md5".
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/gogo/protobuf/proto"
)

//==================================== Validator =============================================

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

// CheckPreOrderDigest is used to check the correctness of digest
func CheckPreOrderDigest(pre *protos.PreOrder) error {
	digest, err := CalculatePreOrderDigest(pre)
	if err != nil {
		return err
	}
	if digest != pre.Digest {
		return errors.New("digest is not equal")
	}
	return nil
}

func CheckOrderAttemptDigest(attempt *protos.OrderAttempt) bool {
	digest, err := CalculateOrderAttemptDigest(attempt)
	if err != nil {
		return false
	}
	if digest != attempt.Digest {
		return false
	}
	return true
}

// CalculatePreOrderDigest is used to calculate the digest
func CalculatePreOrderDigest(pre *protos.PreOrder) (string, error) {
	payload, err := proto.Marshal(&protos.PreOrder{Author: pre.Author, Sequence: pre.Sequence, CommandList: pre.CommandList, TimestampList: pre.TimestampList, ParentDigest: pre.ParentDigest})
	if err != nil {
		return "", err
	}
	return CalculatePayloadHash(payload, 0), nil
}

func CalculateContentDigest(content *protos.OrderAttemptContent) (string, error) {
	payload, err := proto.Marshal(content)
	if err != nil {
		return "", err
	}
	return CalculatePayloadHash(payload, 0), nil
}

func CalculateOrderAttemptDigest(attempt *protos.OrderAttempt) (string, error) {
	payload, err := proto.Marshal(&protos.OrderAttempt{
		NodeID:        attempt.NodeID,
		SeqNo:         attempt.SeqNo,
		ParentDigest:  attempt.ParentDigest,
		ContentDigest: attempt.ContentDigest,
	})
	if err != nil {
		return "", err
	}
	return CalculatePayloadHash(payload, 0), nil
}

// GetHash returns the TransactionHash
func GetHash(tx *protos.Transaction) string {
	if tx.Hash == "" {
		tx.Hash = CalculatePayloadHash(tx.Payload, 0)
	}
	return tx.Hash
}

func CalculateListHash(list []string, timestamp int64) string {
	h := sha256.New()
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

func CalculatePayloadHash(payload []byte, timestamp int64) string {
	return BytesToString(CalculateSHA256Hash(payload, timestamp))
}

func CalculateSHA256Hash(payload []byte, timestamp int64) Hash {
	h := sha256.New()
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
