package external

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
)

type CryptoService interface {
	PrivateKey
	PublicKeys
}

// PrivateKey is an unspecified signature scheme private key
type PrivateKey interface {
	// Algorithm returns the signing algorithm related to the private key.
	Algorithm() string

	// Sign generates a signature using the provided hasher.
	Sign(types.Hash) (*protos.Certification, error)

	// PublicKey returns the public key.
	PublicKey() PublicKey
}

type PublicKeys interface {
	// Algorithm returns the signing algorithm related to the public key.
	Algorithm() string

	// Verify verifies a signature of an input message using the provided hasher with the corresponding public key.
	Verify(uint64, *protos.Certification, types.Hash) error
}

// PublicKey is an unspecified signature scheme public key.
type PublicKey interface {
	// Algorithm returns the signing algorithm related to the public key.
	Algorithm() string

	// Verify verifies a signature of an input message using the provided hasher.
	Verify(*protos.Certification, types.Hash) error
}
