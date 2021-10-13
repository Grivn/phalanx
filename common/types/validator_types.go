package types

import (
	"bytes"
	"encoding/hex"
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
