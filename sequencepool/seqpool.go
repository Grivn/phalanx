package sequencepool

import (
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"sync"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type sequencePool struct {
	// quorum indicates the legal size for stable-state.
	quorum int

	// sts would store the proof for each node.
	sts map[uint64]SyncTree

	// lockedQCs would store the stable-QCs which have been proposed.
	lockedQCs map[uint64]SyncTree

	// commands would store the command we received.
	commands sync.Map
}

// InsertQuorumCert could insert the quorum-cert into sync-tree for specific node.
func (sp *sequencePool) InsertQuorumCert(qc *protos.QuorumCert) {
	sp.sts[qc.Author()].Insert(qc)
}

// InsertCommand could insert command into the sync-map.
func (sp *sequencePool) InsertCommand(command *protos.Command) {
	sp.commands.Store(command.Digest, command)
}

// VerifyQCs is used to verify the QCs in qc-batch.
// 1) we should have quorum QCs in such a batch.
// 2) the qc should contain the specific command for it.
// 3) the sequence number for qc should be matched with the local record for logs of replicas.
// 4) the proof-certs should be valid.
func (sp *sequencePool) VerifyQCs(payload []byte) error {
	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	for _, filter := range qcb.Filters {
		if len(filter.QCs) < sp.quorum {
			return errors.New("not enough qc")
		}

		for _, qc := range filter.QCs {
			if _, ok := qcb.Commands[qc.Digest()]; !ok {
				return errors.New("nil command")
			}

			if sp.sts[qc.Author()].Min().Sequence() != qc.Sequence() {
				return errors.New("invalid sequence number")
			}

			if err := crypto.VerifyProofCerts(types.StringToBytes(qc.Digest()), qc.ProofCerts, sp.quorum); err != nil {
				return fmt.Errorf("verify quourm cert failed: %s", err)
			}
		}
	}

	return nil
}

// LockQCs is used to delete the stable-QCs which have been proposed by other nodes.
func (sp *sequencePool) LockQCs(payload []byte) error {
	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	for _, filter := range qcb.Filters {
		for _, qc := range filter.QCs {
			sp.sts[qc.Author()].Delete(qc)
		}
	}
	return nil
}

// PullQCs is used to pull the QCs from sync-tree to generate consensus proposal.
func (sp *sequencePool) PullQCs() ([]byte, error) {
	qcb := protos.NewQCBatch()
	var qcs []*protos.QuorumCert

	for _, st := range sp.sts {
		qc := st.PullMin()

		// blank:
		// cannot find QC info, continue for next replica log.
		if qc == nil {
			continue
		}

		// state-QC:
		// current QC has been proposed by others and it has reached stable state.
		if sp.lockedQCs[qc.Author()].Has(qc) {
			continue
		}

		// command:
		// we should find the command the QC refers to.
		if _, ok := qcb.Commands[qc.Digest()]; !ok {
			if command := sp.getCommand(qc.Digest()); command == nil {
				continue
			} else {
				qcb.Commands[qc.Digest()] = command
			}
		}

		// append:
		// we have found a QC which could be proposed in next phase, append into QCs slice.
		qcs = append(qcs, qc)
	}

	for _, filter := range qcb.Filters {
		for _, qc := range filter.QCs {
			sp.sts[qc.Author()].Insert(qc)
		}
	}

	if len(qcs) < sp.quorum {
		// there are not enough QCs for current QC
		return nil, errors.New("")
	}

	qcb.Filters = append(qcb.Filters, &protos.QCFilter{QCs: qcs})
	payload, err := proto.Marshal(qcb)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (sp *sequencePool) getCommand(digest string) *protos.Command {
	command, ok := sp.commands.Load(digest)
	if !ok {
		return nil
	}
	return command.(*protos.Command)
}
