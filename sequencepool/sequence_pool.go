package sequencepool

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type sequencePool struct {
	// mutex is used to deal with the concurrent problems of sequence-pool.
	mutex sync.Mutex

	// author indicates the identifier for current participate.
	author uint64

	// reminders would store the proof for each node.
	reminders map[uint64]*partialReminder

	// commands would store the command we received.
	commands map[string]*protos.Command

	// tracker is used to track the duplicated ordered logs for proPartialsal generation.
	tracker CommandTracker

	// rotation indicates the expected window for one block generation in synchronous consensus.
	rotation int

	// duration indicates the network quality in synchronous consensus.
	duration time.Duration

	// isRestored indicates sequence-pool status, and we won't initiate reminder with priori data if it has been restored.
	isRestored bool

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencePool(author uint64, n int, rotation int, duration time.Duration, logger external.Logger) *sequencePool {
	reminders := make(map[uint64]*partialReminder)

	for i:=0; i<n; i++ {
		reminders[uint64(i+1)] = newPartialReminder(author, n, uint64(i+1))
	}

	return &sequencePool{
		author:     author,
		reminders:  reminders,
		commands:   make(map[string]*protos.Command),
		tracker:    NewCommandTracker(n),
		rotation:   rotation,
		duration:   duration,
		isRestored: false,
		logger:     logger,
	}
}

// InsertPartialOrder could insertPartial the quorum-cert into sync-tree for specific node.
func (sp *sequencePool) InsertPartialOrder(pOrder *protos.PartialOrder) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	return sp.reminders[pOrder.Author()].insertPartial(pOrder)
}

// InsertCommand could insertPartial command into the sync-map.
func (sp *sequencePool) InsertCommand(command *protos.Command) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.commands[command.Digest] = command
}

// RestorePartials is used to prepare the status of validator of Partials.
func (sp *sequencePool) RestorePartials() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, reminder := range sp.reminders {
		// restore the Partials in each reminder.
		reminder.restorePartials(sp.tracker)
	}

	// open restored status, which means current cluster has just found invalid partial batch,
	// the priori-batch might be untrusted.
	sp.isRestored = true
}

// VerifyPartials is used to verify the Partials in partial order batch.
// 1) we should have quorum Partials in such a batch.
// 2) the partial order should contain the specific command for it.
// 3) the sequence number for partial order should be matched with the local record for logs of replicas.
// 4) the proof-certs should be valid.
func (sp *sequencePool) VerifyPartials(pBatch *protos.PartialOrderBatch) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if pBatch == nil {
		sp.logger.Infof("verify partial batch: nil partial order")
		return nil
	}

	// verify the validation
	for _, pOrder := range pBatch.Partials {
		if sp.tracker.IsQuorum(pOrder.CommandDigest()) {
			continue
		}

		if _, ok := pBatch.Commands[pOrder.CommandDigest()]; !ok {
			return fmt.Errorf("nil command: replica %d, seqNo %d, digest %s", pOrder.Author(), pOrder.Sequence(), pOrder.CommandDigest())
		}

		if err := sp.reminders[pOrder.Author()].verify(pBatch.Author, pOrder); err != nil {
			return fmt.Errorf("verify partial order failed: %s", err)
		}
	}

	// proposed target
	for _, pOrder := range pBatch.Partials {
		sp.reminders[pOrder.Author()].proposedPartial(pOrder)

		sp.tracker.Add(pOrder.CommandDigest())
	}

	return nil
}

func (sp *sequencePool) SetStablePartials(pBatch *protos.PartialOrderBatch) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if pBatch == nil {
		sp.logger.Infof("[%d] set stable nil batch, ignore it", sp.author)
		return nil
	}

	if len(pBatch.Partials) == 0 {
		// we receive a blank partial batch, which means the priori batch could be trusted,
		// so we can close the restored status and generate batch according to the priori one.
		sp.isRestored = false
		sp.logger.Infof("[%d] set stable a blank partial batch, close restored status", sp.author)
		return nil
	}

	for _, pOrder := range pBatch.Partials {
		sp.logger.Debugf("[%d] set stable %s", sp.author, pOrder.Format())
		if err := sp.reminders[pOrder.Author()].setStablePartial(pOrder); err != nil {
			return fmt.Errorf("stable partial order failed: %s", err)
		}
	}

	return nil
}

// PullPartials is used to pull the Partials from b-tree to generate consensus proposal.
func (sp *sequencePool) PullPartials(priori *protos.PartialOrderBatch) (*protos.PartialOrderBatch, error) {
	time.Sleep(sp.duration)

	sp.mutex.Lock()
	defer sp.mutex.Unlock()
	pBatch := protos.NewPartialOrderBatch(sp.author)

	if sp.isRestored {
		sp.logger.Infof("[%d] just restored sequence-pool, generate a blank priori-batch", sp.author)
		return pBatch, nil
	}

	if priori == nil {
		sp.logger.Debugf("[%d] do not have a priori-batch, trying to generate genesis batch", sp.author)
	} else {
		if len(priori.Partials) == 0 {
			sp.logger.Debugf("[%d] have a blank priori-batch, trying to generate self-dependent batch", sp.author)
		} else {
			// initiate the reminder to avoid duplicated partial orders.
			for id, reminder := range sp.reminders {
				proposedNo := priori.ProposedNos[id]
				reminder.pullInitiation(proposedNo)
				pBatch.ProposedNos[id] = proposedNo
				sp.logger.Debugf("[%d] initiate reminder status, replica %d, proposedNo %d", sp.author, id, proposedNo)
			}
		}
	}

	for i:=0; i<sp.rotation; i++ {
		for _, reminder := range sp.reminders {
			for {
				pOrder := reminder.pullPartial()

				// existence of partial order:
				// cannot find partial order info, continue for next replica's partial order.
				if pOrder == nil {
					break
				}

				// redundancy of partial order:
				// collect the redundant partial order directly for batch generation.
				if sp.tracker.IsQuorum(pOrder.CommandDigest()) {
					pBatch.Append(pOrder)
					sp.logger.Infof("[%d] collect partial order %s", sp.author, pOrder.Format())
					continue
				}

				// existence of command:
				// 1) try to find the command of current partial order in partial batch.
				if _, ok := pBatch.Commands[pOrder.CommandDigest()]; ok {
					pBatch.Append(pOrder)
					sp.logger.Infof("[%d] collect partial order %s", sp.author, pOrder.Format())
					break
				}
				// 2) try to find the command of current partial order in local command reminder.
				if command := sp.getCommand(pOrder.CommandDigest()); command == nil {
					// cannot find current command, then put back the pending partial order.
					reminder.backPartial(pOrder)
				} else {
					pBatch.Commands[pOrder.CommandDigest()] = command
					pBatch.Append(pOrder)
					sp.logger.Infof("[%d] collect partial order %s", sp.author, pOrder.Format())
				}
				break
			}
		}
	}

	if len(pBatch.Partials) == 0 {
		// we cannot find any valid partial order the generate batch, return failure message
		return nil, errors.New("failed to generate a batch, no valid partial order")
	}

	// for replica who has generated a partial batch, which should be a trusted batch for itself,
	// before next phalanx-restoring, current node could find a trusted priori-batch,
	// so that we should close the restored status and initiate the reminder with the priori-batch.
	sp.isRestored = false
	return pBatch, nil
}

func (sp *sequencePool) getCommand(digest string) *protos.Command {
	command, ok := sp.commands[digest]
	if !ok {
		return nil
	}
	return command
}
