package sgx

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"fmt"

	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

// USIG implements USIG interface around USIGEnclave.
type usigImpl struct {
	enclave *enclaveImpl
}

// NewUSIG creates a new instance of SGXUSIG. It is a wrapper around
// NewUSIGEnclave(). See NewUSIGEnclave() for more details. Note that
// the created instance has to be disposed with Destroy() method, e.g.
// using defer.
func NewUSIG(enclaveFile string, sealedKey []byte) (*usigImpl, error) {
	enclave, err := NewUSIGEnclave(enclaveFile, sealedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create USIG enclave: %v", err)
	}

	return &usigImpl{enclave: enclave}, nil
}

// CreateUI creates a unique identifier assigned to the message.
func (u *usigImpl) CreateUI(message []byte) (*commonProto.UI, error) {
	counter, signature, err := u.enclave.CreateUI(messageDigest(message))
	if err != nil {
		return nil, err
	}

	return &commonProto.UI{
		Counter: counter,
		Cert:    MakeCert(u.enclave.Epoch(), signature),
	}, nil
}

// VerifyUI is just a wrapper around the VerifyUI function at the
// package-level.
func (u *usigImpl) VerifyUI(message []byte, ui *commonProto.UI, usigID []byte) error {
	return VerifyUI(message, ui, usigID)
}

// ID returns the USIG instance identity.
func (u *usigImpl) ID() []byte {
	id, err := MakeID(u.enclave.Epoch(), u.enclave.PublicKey())
	if err != nil {
		panic(err)
	}
	return id
}

// VerifyUI verifies unique identifier generated for the message by
// USIG with the specified identity.
func VerifyUI(message []byte, ui *commonProto.UI, usigID []byte) error {
	epoch, pubKey, err := ParseID(usigID)
	if err != nil {
		return fmt.Errorf("failed to parse USIG ID: %s", err)
	}

	uiEpoch, signature, err := ParseCert(ui.Cert)
	if err != nil {
		return fmt.Errorf("failed to parse UI cert: %s", err)
	}

	if uiEpoch != epoch {
		return fmt.Errorf("epoch value mismatch")
	}

	return VerifySignature(pubKey, messageDigest(message), epoch, ui.Counter, signature)
}

func messageDigest(message []byte) Digest {
	return sha256.Sum256(message)
}

// MakeID composes a USIG identity which is 64-bit big-endian encoded
// epoch value followed by public key serialized in PKIX format.
func MakeID(epoch uint64, publicKey interface{}) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize public key: %s", err)
	}

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, epoch); err != nil {
		panic(err)
	}

	if err := binary.Write(buf, binary.BigEndian, publicKeyBytes); err != nil {
		panic(err)
	}

	return buf.Bytes(), nil
}

// ParseID breaks a USIG identity down to epoch value and public key.
func ParseID(usigID []byte) (epoch uint64, pubKey crypto.PublicKey, err error) {
	buf := bytes.NewBuffer(usigID)

	err = binary.Read(buf, binary.BigEndian, &epoch)
	if err != nil {
		return uint64(0), nil, fmt.Errorf("failed to extract epoch from USIG ID: %s", err)
	}

	pubKey, err = x509.ParsePKIXPublicKey(buf.Bytes())
	if err != nil {
		return uint64(0), nil, fmt.Errorf("failed to parse public key: %s", err)
	}

	return epoch, pubKey, err
}

// MakeCert composes a USIG certificate which is 64-bit big-endian
// encoded epoch value followed by serialized USIG signature.
func MakeCert(epoch uint64, signature []byte) []byte {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, epoch); err != nil {
		panic(err)
	}

	if err := binary.Write(buf, binary.BigEndian, signature); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// ParseCert breaks a USIG certificate down to epoch value and
// serialized USIG signature.
func ParseCert(cert []byte) (epoch uint64, signature []byte, err error) {
	buf := bytes.NewBuffer(cert)

	err = binary.Read(buf, binary.BigEndian, &epoch)
	if err != nil {
		return uint64(0), nil, fmt.Errorf("failed to extract epoch from USIG cert: %s", err)
	}

	return epoch, buf.Bytes(), nil
}