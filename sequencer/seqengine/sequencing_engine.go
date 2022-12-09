package seqengine

import (
	"github.com/Grivn/phalanx/lib/instance"
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type sequencingEngine struct {
	// mutex is used to handle the concurrency problems.
	mutex sync.Mutex

	// sequencerID is the identifier of current sequencer node.
	sequencerID uint64

	// cIndexC is used to feedback the command_index to trigger the generation of order-attempts.
	cIndexC chan *types.CommandIndex

	// relays are used to track the commands received.
	relays map[uint64]api.Relay

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencingEngine(sequencerID uint64, logger external.Logger) api.SequencingEngine {
	return &sequencingEngine{
		sequencerID: sequencerID,
		cIndexC:     make(chan *types.CommandIndex),
		relays:      make(map[uint64]api.Relay),
		logger:      logger,
	}
}

func (seq *sequencingEngine) Sequencing(command *protos.Command) {
	seq.mutex.Lock()
	defer seq.mutex.Unlock()

	// Select the relay module.
	relay, ok := seq.relays[command.Author]
	if !ok {
		// If there is not a relay instance, initiate it.
		seq.logger.Errorf("[%d] don't have client instance %d, initiate it", seq.sequencerID, command.Author)
		relay = instance.NewRelay(seq.sequencerID, command.Author, seq.cIndexC, seq.logger)
		seq.relays[command.Author] = relay
	}

	// Append the command into this relay module.
	relay.Append(command)
}

func (seq *sequencingEngine) CommandIndexChan() <-chan *types.CommandIndex {
	return seq.cIndexC
}
