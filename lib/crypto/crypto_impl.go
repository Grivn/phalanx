package crypto

import (
	"fmt"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type cryptoImpl struct {
	privateKey external.PrivateKey
	publicKeys map[uint64]external.PublicKey
}

func NewCrypto(privateKey external.PrivateKey, publicKeys map[uint64]external.PublicKey) api.Crypto {
	return &cryptoImpl{privateKey: privateKey, publicKeys: publicKeys}
}

func (c *cryptoImpl) PrivateSign(hash types.Hash) (*protos.Certification, error) {
	return c.privateKey.Sign(hash)
}

func (c *cryptoImpl) PublicVerify(cert *protos.Certification, hash types.Hash, nodeID uint64) error {
	verifier, ok := c.publicKeys[nodeID]
	if !ok {
		return fmt.Errorf("cannot find verifier for node %d", nodeID)
	}
	return verifier.Verify(cert, hash)
}

// VerifyProofCerts is used to verify the validation of proof-certs
func (c *cryptoImpl) VerifyProofCerts(digest types.Hash, pc *protos.QuorumCert, quorum int) error {
	if pc == nil {
		return fmt.Errorf("nil proof-certs")
	}
	if len(pc.Certs) < quorum {
		return fmt.Errorf("not enough signatures, expect %d, received %d", quorum, len(pc.Certs))
	}
	for id, cert := range pc.Certs {
		if err := c.PublicVerify(cert, digest, id); err != nil {
			return fmt.Errorf("illegal cert from node %d: %s", id, err)
		}
	}
	return nil
}
