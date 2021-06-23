package sequencepool

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

type sequencePool struct {
	// mutex is used to deal with the concurrent problems of sequence-pool.
	mutex sync.Mutex

	// author indicates the identifier for current participate.
	author uint64

	// quorum indicates the legal size for stable-state.
	quorum int

	// reminders would store the proof for each node.
	reminders map[uint64]*qcReminder

	// commands would store the command we received.
	commands map[string]*protos.Command

	//
	tracker CommandTracker
}

func NewSequencePool(author uint64, n int) *sequencePool {
	reminders := make(map[uint64]*qcReminder)

	for i:=0; i<n; i++ {
		reminders[uint64(i+1)] = newQCReminder(author, n, uint64(i+1))
	}

	return &sequencePool{author: author, quorum: types.CalculateOneQuorum(n), reminders: reminders, commands: make(map[string]*protos.Command), tracker: NewCommandTracker(n)}
}

func (sp *sequencePool) BecomeLeader() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, reminder := range sp.reminders {
		reminder.becomeLeader()
	}
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

// RestoreQCs is used to prepare the status of validator of QCs.
func (sp *sequencePool) RestoreQCs() {
	for _, reminder := range sp.reminders {
		// restore the QCs in each reminder.
		reminder.restoreQCs(sp.tracker)
	}
}

// VerifyQCs is used to verify the QCs in qc-batch.
// 1) we should have quorum QCs in such a batch.
// 2) the qc should contain the specific command for it.
// 3) the sequence number for qc should be matched with the local record for logs of replicas.
// 4) the proof-certs should be valid.
func (sp *sequencePool) VerifyQCs(qcb *protos.QCBatch) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// verify the validation
	for _, filter := range qcb.Filters {
		if len(filter.QCs) < sp.quorum {
			return errors.New("not enough qc")
		}

		for _, qc := range filter.QCs {
			if sp.tracker.IsQuorum(qc.CommandDigest()) {
				continue
			}

			if _, ok := qcb.Commands[qc.CommandDigest()]; !ok {
				return fmt.Errorf("nil command: replica %d, seqNo %d, digest %s", qc.Author(), qc.Sequence(), qc.CommandDigest())
			}

			if err := sp.reminders[qc.Author()].verify(qcb.Author, qc); err != nil {
				return fmt.Errorf("verify QC failed: %s", err)
			}
		}
	}

	// proposed target
	for _, filter := range qcb.Filters {
		for _, qc := range filter.QCs {
			sp.reminders[qc.Author()].proposedQC(qc)

			sp.tracker.Add(qc.CommandDigest())
		}
	}

	return nil
}

func (sp *sequencePool) SetStableQCs(qcb *protos.QCBatch) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, filter := range qcb.Filters {
		for _, qc := range filter.QCs {
			if err := sp.reminders[qc.Author()].setStableQC(qc); err != nil {
				return fmt.Errorf("stable QC failed: %s", err)
			}
		}
	}

	return nil
}

// PullQCs is used to pull the QCs from sync-tree to generate consensus proposal.
func (sp *sequencePool) PullQCs() (*protos.QCBatch, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	qcb := protos.NewQCBatch(sp.author)
	var qcs []*protos.QuorumCert

	count := 0
	for _, reminder := range sp.reminders {
		var qc *protos.QuorumCert

		for {
			qc = reminder.pullQC()

			// blank:
			// cannot find QC info, continue for next replica log.
			if qc == nil {
				break
			}

			if sp.tracker.NonQuorum(qc.CommandDigest()) {
				break
			}

			qcs = append(qcs, qc)
		}

		if qc == nil {
			break
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

		count++
	}

	// todo pre-generate for block

	if count < sp.quorum {
		for _, qc := range qcs {
			// push the unavailable QCs back
			sp.reminders[qc.Author()].backQC(qc)
		}

		// there are not enough QCs for current QC
		return nil, fmt.Errorf("not enough QCs, need %d, has %d", sp.quorum, len(qcs))
	}

	qcb.Filters = append(qcb.Filters, &protos.QCFilter{QCs: qcs})
	return qcb, nil
}

func (sp *sequencePool) getCommand(digest string) *protos.Command {
	command, ok := sp.commands[digest]
	if !ok {
		return nil
	}
	return command
}
