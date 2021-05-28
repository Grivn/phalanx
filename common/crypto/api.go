package crypto

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

// PrivateKey is an unspecified signature scheme private key
type PrivateKey interface {
	// Algorithm returns the signing algorithm related to the private key.
	Algorithm() string

	// Sign generates a signature using the provided hasher.
	Sign(types.Hash) (*protos.Certification, error)

	// PublicKey returns the public key.
	PublicKey() PublicKey
}

// PublicKey is an unspecified signature scheme public key.
type PublicKey interface {
	// Algorithm returns the signing algorithm related to the public key.
	Algorithm() string

	// Verify verifies a signature of an input message using the provided hasher.
	Verify(*protos.Certification, types.Hash) error
}

// Hasher interface
type Hasher interface {
	// ComputeHash returns the hash output regardless of the hash state
	ComputeHash([]byte) types.Hash
}
