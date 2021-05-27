package crypto

import (
	"hash"

	"github.com/Grivn/phalanx/common/types"

	"golang.org/x/crypto/sha3"
)

const (
	// SHA3256LEN is the length for hash value of sha3 256
	SHA3256LEN = 32
)

// MakeID creates an ID from the hash of encoded data.
func MakeID(payload []byte) types.Identifier {
	return HashToID(NewSHA3256().ComputeHash(payload))
}

func HashToID(hash []byte) types.Identifier {
	var id types.Identifier
	copy(id[:], hash)
	return id
}

func IDToByte(id types.Identifier) []byte {
	return id[:]
}

//================================= SHA3 256 ============================================

// sha3256Algo is an implement for hasher with sha3 256
type sha3256Algo struct {
	size int
	hash.Hash
}

// NewSHA3256 returns a new instance of SHA3-256 hasher
func NewSHA3256() Hasher {
	return &sha3256Algo{
		size: SHA3256LEN,
		Hash: sha3.New256(),
	}
}

// ComputeHash calculates and returns the SHA3-256 output of input byte array
func (s *sha3256Algo) ComputeHash(data []byte) types.Hash {
	s.Reset()
	_, _ = s.Write(data)
	digest := make(types.Hash, 0, SHA3256LEN)
	return s.Sum(digest)
}
