package sequencing

import (
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/lib/instance"
)

type sequencingEngine struct {
	// mutex is used to handle the concurrency problems.
	mutex sync.Mutex

	// sequencerID is the identifier of current sequencer node.
	sequencerID uint64

	// eventC is used to feedback the command_index to trigger the generation of order-attempts.
	eventC chan types.LocalEvent

	// relays are used to track the commands received.
	relays map[uint64]api.Relay

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencingEngine(sequencerID uint64, eventC chan types.LocalEvent, logger external.Logger) api.SequencingEngine {
	return &sequencingEngine{
		sequencerID: sequencerID,
		eventC:      eventC,
		relays:      make(map[uint64]api.Relay),
		logger:      logger,
	}
}

func (engine *sequencingEngine) Sequencing(command *protos.Command) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	// Select the relay module.
	module, ok := engine.relays[command.Author]
	if !ok {
		// If there is not a relay instance, initiate it.
		engine.logger.Errorf("[%d] don't have client instance %d, initiate it", engine.sequencerID, command.Author)
		module = instance.NewRelay(engine.sequencerID, command.Author, engine.eventC, engine.logger)
		engine.relays[command.Author] = module
	}

	// Append the command into this relay module.
	module.Append(command)
}
