package types

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"github.com/Grivn/phalanx/common/types/protos"

	fTypes "github.com/ultramesh/flato-common/types"
)

// GetHash returns the TransactionHash
func GetHash(tx *protos.Transaction) string {
	return fTypes.GetHash(tx.Tx).Hex()
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

func CalculateMD5Hash(list []string, timestamp int64) []byte {
	h := md5.New()

	for _, hash := range list {
		_, _ = h.Write([]byte(hash))
	}

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}

	return h.Sum(nil)
}
