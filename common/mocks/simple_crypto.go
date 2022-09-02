package mocks

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/Grivn/phalanx/external"
	"math/big"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

const (
	COUNT = 512
	// ECDSA_P256 is supported for crypto algorithms
	ECDSA_P256 = "ECDSA_P256"
)

//==================================== create naive validator =============================================

type StaticRand struct{ id int }

func (sr *StaticRand) Read(x []byte) (int, error) { return sr.id, nil }

// GenerateKeys is used to init the public/private keys for validator
func GenerateKeys(author uint64, count int) (external.PrivateKey, map[uint64]external.PublicKey, error) {
	var privKey external.PrivateKey
	pubKeys := make(map[uint64]external.PublicKey, count)
	for i := 0; i < count; i++ {
		id := uint64(i + 1)
		pair, err := generateKeyHelper(ECDSA_P256, i+1)
		if err != nil {
			return nil, nil, err
		}
		pubKeys[id] = pair.PublicKey()
		if id == author {
			privKey = pair
		}
	}
	return privKey, pubKeys, nil
}

func generateKeyHelper(signer string, id int) (external.PrivateKey, error) {
	if signer == ECDSA_P256 {
		// generate key pairs with static id
		pubkeyCurve := elliptic.P256()
		priv, err := ecdsa.GenerateKey(pubkeyCurve, &StaticRand{id: id})
		if err != nil {
			return nil, err
		}
		privKey := &ecdsaP256PrivateKey{SignAlg: signer, PrivateKey: priv}
		return privKey, nil
	}
	return nil, fmt.Errorf("invalid signature scheme %s", signer)
}

// ================================= ecdsa implementation ====================================

type ecdsaSignature struct {
	r, s *big.Int
}

type ecdsaP256PrivateKey struct {
	SignAlg    string
	PrivateKey *ecdsa.PrivateKey
}

type ecdsaP256PublicKey struct {
	SignAlg   string
	PublicKey ecdsa.PublicKey
}

// PublicKey returns the public key.
func (priv *ecdsaP256PrivateKey) PublicKey() external.PublicKey {
	pub := &ecdsaP256PublicKey{SignAlg: ECDSA_P256, PublicKey: priv.PrivateKey.PublicKey}
	return pub
}

// Algorithm returns the signing algorithm related to the private key.
func (priv *ecdsaP256PrivateKey) Algorithm() string {
	return priv.SignAlg
}

// Sign returns two *big.Int variables. In order to save it in the Signature type,
// I first turn them into strings, and then I turn the strings to byte arrays.
// I have implemented a Signature to ecdsa signature parser (toECDSA in signature.go) in oder to
// cast the byte array Signature into the original signature of the ECDSA signing method.
func (priv *ecdsaP256PrivateKey) Sign(hash types.Hash) (*protos.Certification, error) {
	var r, s *big.Int
	var err error
	r, s, err = ecdsa.Sign(rand.Reader, priv.PrivateKey, hash)
	if err != nil {
		return nil, err
	}
	sig := make([][]byte, 2)
	sig[0] = []byte(r.String())
	sig[1] = []byte(s.String())
	return &protos.Certification{Signatures: sig}, err
}

// Algorithm returns the signing algorithm related to the public key.
func (pub *ecdsaP256PublicKey) Algorithm() string {
	return pub.SignAlg
}

// Verify verifies a signature of an input message using the provided hasher.
func (pub *ecdsaP256PublicKey) Verify(cert *protos.Certification, hash types.Hash) error {
	ecdsaSig := signatureToECDSA(cert)
	if pass := ecdsa.Verify(&pub.PublicKey, hash, ecdsaSig.r, ecdsaSig.s); pass {
		return nil
	}
	return errors.New("invalid signature")
}

func signatureToECDSA(cert *protos.Certification) ecdsaSignature {
	var ecdsaSig = new(ecdsaSignature)
	var r, s big.Int
	r.SetString(string(cert.Signatures[0]), 10)
	s.SetString(string(cert.Signatures[1]), 10)
	ecdsaSig.r = &r
	ecdsaSig.s = &s
	return *ecdsaSig
}
