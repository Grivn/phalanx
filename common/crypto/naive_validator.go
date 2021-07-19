package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"

	"github.com/gogo/protobuf/proto"
)

//==================================== Naive Validator =============================================

var keys []PrivateKey
var pubKeys []PublicKey

// SetKeys is used to init the public/private keys for validator
func SetKeys() error {
	keys = make([]PrivateKey, types.COUNT)
	pubKeys = make([]PublicKey, types.COUNT)
	var err error
	for i := 0; i < types.COUNT; i++ {
		keys[i], err = generateKey(types.ECDSA_P256, i+1)
		if err != nil {
			return err
		}
		pubKeys[i] = keys[i].PublicKey()
	}
	return nil
}

// PrivSign is used to generate signature with private key
func PrivSign(hash types.Hash, nodeID int) (*protos.Certification, error) {
	return keys[nodeID-1].Sign(hash)
}

// PubVerify is used to verify the signature with the provided public key
func PubVerify(cert *protos.Certification, hash types.Hash, nodeID int) error {
	return pubKeys[nodeID-1].Verify(cert, hash)
}

// VerifyProofCerts is used to verify the validation of proof-certs
func VerifyProofCerts(digest types.Hash, pc *protos.PQuorumCert, quorum int) error {
	if pc == nil {
		return errors.New("nil proof-certs")
	}
	if len(pc.Certs) < quorum {
		return errors.New("not enough signatures")
	}
	for id, cert := range pc.Certs {
		if err := PubVerify(cert, digest, int(id)); err != nil {
			return err
		}
	}
	return nil
}

// CheckDigest is used to check the correctness of digest
func CheckDigest(pre *protos.PreOrder) error {
	payload, err := proto.Marshal(&protos.PreOrder{Author: pre.Author, Sequence: pre.Sequence, CommandDigest: pre.CommandDigest, Timestamp: pre.Timestamp})
	if err != nil {
		return err
	}
	if types.CalculatePayloadHash(payload, 0) != pre.Digest {
		return errors.New("digest is not equal")
	}
	return nil
}

//==================================== Helper =============================================

type StaticRand struct { id int }
func (sr *StaticRand) Read(x []byte) (int, error) { return sr.id, nil }
func generateKey(signer string, id int) (PrivateKey, error) {
	if signer == types.ECDSA_P256 {
		pubkeyCurve := elliptic.P256()
		// use static id
		priv, err := ecdsa.GenerateKey(pubkeyCurve, &StaticRand{id: id})
		if err != nil {
			return nil, err
		}
		privKey := &ecdsaP256PrivateKey{SignAlg: signer, PrivateKey: priv}
		return privKey, nil
	} else if signer == types.ECDSA_SECp256k1 {
		return nil, nil
	} else if signer == types.BLS_BLS12381 {
		return nil, nil
	} else {
		return nil, errors.New("invalid signature scheme")
	}
}
