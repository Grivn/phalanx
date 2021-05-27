package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"math/big"
)

// ================================= ECDSA ====================================

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
func (priv *ecdsaP256PrivateKey) PublicKey() PublicKey {
	pub := &ecdsaP256PublicKey{SignAlg: types.ECDSA_P256, PublicKey: priv.PrivateKey.PublicKey}
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
func (pub *ecdsaP256PublicKey) Verify(cert *protos.Certification, hash types.Hash) (bool, error) {
	ecdsaSig := signatureToECDSA(cert)
	isVerified := ecdsa.Verify(&pub.PublicKey, hash, ecdsaSig.r, ecdsaSig.s)
	return isVerified, nil
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
