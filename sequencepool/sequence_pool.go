package sequencepool

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"

	"github.com/gogo/protobuf/proto"
)

type sequencePool struct {
	// mutex is used to deal with the concurrent problems of sequence-pool.
	mutex sync.Mutex

	// quorum indicates the legal size for stable-state.
	quorum int

	// reminders would store the proof for each node.
	reminders map[uint64]*qcReminder

	// stableS indicates the stable-sequence for each participate.
	stableS map[uint64]uint64

	// commands would store the command we received.
	commands map[string]*protos.Command
}

func NewSequencePool(author uint64, n int) *sequencePool {
	reminders := make(map[uint64]*qcReminder)

	for i:=0; i<n; i++ {
		reminders[uint64(i+1)] = newQCReminder(author, n, uint64(i+1))
	}

	return &sequencePool{quorum: types.CalculateQuorum(n), reminders: reminders, commands: make(map[string]*protos.Command)}
}

// InsertQuorumCert could insertQC the quorum-cert into sync-tree for specific node.
func (sp *sequencePool) InsertQuorumCert(qc *protos.QuorumCert) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	return sp.reminders[qc.Author()].insertQC(qc)
}

// InsertCommand could insertQC command into the sync-map.
func (sp *sequencePool) InsertCommand(command *protos.Command) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.commands[command.Digest] = command
}

// VerifyQCs is used to verify the QCs in qc-batch.
// 1) we should have quorum QCs in such a batch.
// 2) the qc should contain the specific command for it.
// 3) the sequence number for qc should be matched with the local record for logs of replicas.
// 4) the proof-certs should be valid.
func (sp *sequencePool) VerifyQCs(payload []byte) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	// init the reminder for each participate before the verification for QCs
	for _, reminder := range sp.reminders {
		reminder.preprocess()
	}

	for _, filter := range qcb.Filters {
		if len(filter.QCs) < sp.quorum {
			return errors.New("not enough qc")
		}

		for _, qc := range filter.QCs {
			if _, ok := qcb.Commands[qc.CommandDigest()]; !ok {
				return errors.New("nil command")
			}

			if err := sp.reminders[qc.Author()].verify(qc); err != nil {
				return fmt.Errorf("verify QC failed: %s", err)
			}

			sp.reminders[qc.Author()].lockQC(qc)
		}
	}

	return nil
}

func (sp *sequencePool) StableQCs(payload []byte) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	qcb := &protos.QCBatch{}
	if err := proto.Unmarshal(payload, qcb); err != nil {
		return fmt.Errorf("invalid QC-batch: %s", err)
	}

	for _, filter := range qcb.Filters {
		for _, qc := range filter.QCs {
			if err := sp.reminders[qc.Author()].stableQC(qc); err != nil {
				return fmt.Errorf("stable QC failed: %s", err)
			}
		}
	}

	return nil
}

// PullQCs is used to pull the QCs from sync-tree to generate consensus proposal.
func (sp *sequencePool) PullQCs() ([]byte, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	qcb := protos.NewQCBatch()
	var qcs []*protos.QuorumCert

	for _, reminder := range sp.reminders {
		qc := reminder.pullQC()

		// blank:
		// cannot find QC info, continue for next replica log.
		if qc == nil {
			continue
		}

		// command:
		// we should find the command the QC refers to.
		if _, ok := qcb.Commands[qc.CommandDigest()]; !ok {
			if command := sp.getCommand(qc.CommandDigest()); command == nil {
				//fmt.Printf("don't have %s\n", qc.Digest())
				continue
			} else {
				qcb.Commands[qc.CommandDigest()] = command
			}
		}

		// append:
		// we have found a QC which could be proposed in next phase, append into QCs slice.
		qcs = append(qcs, qc)
	}

	if len(qcs) < sp.quorum {
		// there are not enough QCs for current QC
		return nil, fmt.Errorf("not enough QCs, need %d, has %d", sp.quorum, len(qcs))
	}

	qcb.Filters = append(qcb.Filters, &protos.QCFilter{QCs: qcs})
	payload, err := proto.Marshal(qcb)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (sp *sequencePool) getCommand(digest string) *protos.Command {
	command, ok := sp.commands[digest]
	if !ok {
		return nil
	}
	return command
}
