package crypto

import (
	"fmt"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
)

type cryptoService struct {
	privateKey external.PrivateKey
	publicKeys external.PublicKeys
}

func NewCryptoService(privateKey external.PrivateKey, publicKeys external.PublicKeys) api.CryptoService {
	return &cryptoService{privateKey: privateKey, publicKeys: publicKeys}
}

func (c *cryptoService) PrivateSign(hash types.Hash) (*protos.Certification, error) {
	return c.privateKey.Sign(hash)
}

func (c *cryptoService) PublicVerify(cert *protos.Certification, hash types.Hash, nodeID uint64) error {
	return c.publicKeys.Verify(nodeID, cert, hash)
}

// VerifyProofCerts is used to verify the validation of proof-certs
func (c *cryptoService) VerifyProofCerts(digest types.Hash, pc *protos.QuorumCert, quorum int) error {
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
