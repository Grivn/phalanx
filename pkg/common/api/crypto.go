package api

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
)

type CryptoService interface {
	Signer
	Verifier
}

type Signer interface {
	PrivateSign(hash types.Hash) (*protos.Certification, error)
}

type Verifier interface {
	PublicVerify(cert *protos.Certification, hash types.Hash, nodeID uint64) error
	VerifyProofCerts(digest types.Hash, pc *protos.QuorumCert, quorum int) error
}
