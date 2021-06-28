package sequencepool

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type sequencePool struct {
	// mutex is used to deal with the concurrent problems of sequence-pool.
	mutex sync.Mutex

	// author indicates the identifier for current participate.
	author uint64

	// oneQuorum indicates the legal size for stable-state.
	oneQuorum int

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

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencePool(author uint64, n int, rotation int, duration time.Duration, logger external.Logger) *sequencePool {
	reminders := make(map[uint64]*partialReminder)

	for i:=0; i<n; i++ {
		reminders[uint64(i+1)] = newPartialReminder(author, n, uint64(i+1))
	}

	return &sequencePool{
		author:    author,
		oneQuorum: types.CalculateOneQuorum(n),
		reminders: reminders,
		commands:  make(map[string]*protos.Command),
		tracker:   NewCommandTracker(n),
		rotation:  rotation,
		duration:  duration,
		logger:    logger,
	}
}

func (sp *sequencePool) BecomeLeader() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, reminder := range sp.reminders {
		reminder.becomeLeader()
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
	for _, reminder := range sp.reminders {
		// restore the Partials in each reminder.
		reminder.restorePartials(sp.tracker)
	}
}

// VerifyPartials is used to verify the Partials in partial order batch.
// 1) we should have quorum Partials in such a batch.
// 2) the partial order should contain the specific command for it.
// 3) the sequence number for partial order should be matched with the local record for logs of replicas.
// 4) the proof-certs should be valid.
func (sp *sequencePool) VerifyPartials(batch *protos.PartialOrderBatch) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// verify the validation
	for _, filter := range batch.PartialSet {
		//if len(filter.PartialOrders) < sp.oneQuorum {
		//	return errors.New("not enough partial order")
		//}

		for _, pOrder := range filter.PartialOrders {
			if sp.tracker.IsQuorum(pOrder.CommandDigest()) {
				continue
			}

			if _, ok := batch.Commands[pOrder.CommandDigest()]; !ok {
				return fmt.Errorf("nil command: replica %d, seqNo %d, digest %s", pOrder.Author(), pOrder.Sequence(), pOrder.CommandDigest())
			}

			if err := sp.reminders[pOrder.Author()].verify(batch.Author, pOrder); err != nil {
				return fmt.Errorf("verify partial order failed: %s", err)
			}
		}
	}

	// proposed target
	for _, filter := range batch.PartialSet {
		for _, pOrder := range filter.GetPartialOrders() {
			sp.reminders[pOrder.Author()].proposedPartial(pOrder)

			sp.tracker.Add(pOrder.CommandDigest())
		}
	}

	return nil
}

func (sp *sequencePool) SetStablePartials(batch *protos.PartialOrderBatch) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, filter := range batch.PartialSet {
		for _, pOrder := range filter.PartialOrders {
			if err := sp.reminders[pOrder.Author()].setStablePartial(pOrder); err != nil {
				return fmt.Errorf("stable partial order failed: %s", err)
			}
		}
	}

	return nil
}

// PullPartials is used to pull the Partials from sync-tree to generate consensus proPartialsal.
func (sp *sequencePool) PullPartials() (*protos.PartialOrderBatch, error) {
	time.Sleep(sp.duration)

	sp.mutex.Lock()
	defer sp.mutex.Unlock()
	batch := protos.NewPartialOrderBatch(sp.author)

	for i:=0; i<sp.rotation; i++ {
		var pOrderSet []*protos.PartialOrder
		count := 0
		for _, reminder := range sp.reminders {
			var pOrder *protos.PartialOrder
			var tmpPartials []*protos.PartialOrder
			for {
				pOrder = reminder.pullPartial()

				// blank:
				// cannot find partial order info, continue for next replica log.
				if pOrder == nil {
					break
				}
				tmpPartials = append(tmpPartials, pOrder)

				if sp.tracker.NonQuorum(pOrder.CommandDigest()) {
					break
				}
			}

			if pOrder == nil {
				for _, tmpPO := range tmpPartials {
					reminder.backPartial(tmpPO)
				}
				continue
			}

			// command:
			// we should find the command the partial order refers to.
			if _, ok := batch.Commands[pOrder.CommandDigest()]; !ok {
				if command := sp.getCommand(pOrder.CommandDigest()); command == nil {
					for _, tmpPO := range tmpPartials {
						reminder.backPartial(tmpPO)
					}
					continue
				} else {
					batch.Commands[pOrder.CommandDigest()] = command
				}
			}

			// append:
			// we have found a partial order which could be proPartialsed in next phase, append into Partials slice.
			pOrderSet = append(pOrderSet, tmpPartials...)

			count++
		}

		//if count < sp.oneQuorum {
		//	// there are not enough Partials for current partial order
		//	// oneQuorum here (f+1) indicates that there is at least one correct node has finished selfish order and
		//	// trying to trigger consensus phase.
		//	for _, pOrder := range pOrderSet {
		//		// push the unavailable Partials back
		//		sp.reminders[pOrder.Author()].backPartial(pOrder)
		//	}
		//	break
		//}

		batch.PartialSet = append(batch.PartialSet, &protos.PartialSet{PartialOrders: pOrderSet})
	}

	if len(batch.PartialSet) == 0 {
		return nil, errors.New("failed to generate a proPartialsal")
	}

	return batch, nil
}

func (sp *sequencePool) getCommand(digest string) *protos.Command {
	command, ok := sp.commands[digest]
	if !ok {
		return nil
	}
	return command
}
