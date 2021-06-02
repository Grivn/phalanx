package protos

import (
	"errors"
	"fmt"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/types"
	"github.com/gogo/protobuf/proto"
)

func (m *Proposal) Format() string {
	return fmt.Sprintf("[Proposal, Author %d, Sequence %d, Batch %s]", m.Author, m.Sequence, m.TxBatch.Digest)
}

func (m *PreOrder) CheckDigest() error {
	payload, err := proto.Marshal(&PreOrder{Author: m.Author, Sequence: m.Sequence, BatchDigest: m.BatchDigest, Timestamp: m.Timestamp})
	if err != nil {
		return err
	}
	if types.CalculatePayloadHash(payload, 0) != m.Digest {
		return errors.New("digest is not equal")
	}
	return nil
}

func (m *Order) Digest() string {
	return m.PreOrder.Digest
}

func (m *Order) Verify(quorum int) error {
	if len(m.ProofCerts) < quorum {
		return errors.New("not enough signatures")
	}
	for id, cert := range m.ProofCerts {
		if err := crypto.PubVerify(cert, types.StringToBytes(m.Digest()), int(id)); err != nil {
			return err
		}
	}
	return nil
}