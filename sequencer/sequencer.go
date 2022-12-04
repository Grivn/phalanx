package sequencer

import (
	"fmt"
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/lib/utils"
	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/sequencer/sequencing"
	"sort"
)

type sequencerImpl struct {
	//===================================== basic information =========================================

	// author is the identifier for current node.
	sequencerID uint64

	// byz indicates if current node is the adversary.
	byz bool

	//==================================== sub-chain management =============================================

	//
	highestAttempt *protos.OrderAttempt

	// commandSet is used to record the commands' waiting list according to receive order.
	commandSet types.CommandSet

	//===================================== client commands manager ============================================

	// eventC is used to receive local_event from other modules.
	eventC chan types.LocalEvent

	// closeC is used to stop log manager.
	closeC chan bool

	//=================================== local timer service ========================================

	// timer is used to control the timeout event to generate order with commands in waiting list.
	timer api.LocalTimer

	// timeoutC is used to receive timeout event.
	timeoutC <-chan bool

	// engine is used to sequencing the commands received and create command_index to propose order-attempts.
	engine api.SequencingEngine

	//======================================= external tools ===========================================

	// sender is used to send consensus message into network.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger

	// metrics is used to record the metric info of current node's meta pool.
	metrics *metrics.MetaPoolMetrics
}

func NewSequencer(conf Config) *sequencerImpl {
	conf.Logger.Infof("[%d] initiate log manager, replica count %d", conf.Author, conf.N)

	// initiate communication channel.
	eventC := make(chan types.LocalEvent)
	timeoutC := make(chan bool)

	return &sequencerImpl{
		sequencerID: conf.Author,
		timer:       utils.NewLocalTimer(conf.Author, timeoutC, conf.Duration, conf.Logger),
		eventC:      eventC,
		timeoutC:    timeoutC,
		closeC:      make(chan bool),
		engine:      sequencing.NewSequencingEngine(conf.Author, eventC, conf.Logger),
		sender:      conf.Sender,
		logger:      conf.Logger,
		metrics:     conf.Metrics,
		byz:         conf.Byz,
	}
}

func (ser *sequencerImpl) Run() {
	for {
		select {
		case <-ser.closeC:
			return
		case ev := <-ser.eventC:
			ser.dispatchLocalEvent(ev)
		case <-ser.timeoutC:
			if err := ser.generateOrderAttempt(); err != nil {
				panic(fmt.Sprintf("log manager runtime error: %s", err))
			}
		}
	}
}

func (ser *sequencerImpl) dispatchLocalEvent(event types.LocalEvent) {
	switch event.Type {
	case types.LocalEventCommand:
		command, ok := event.Event.(*protos.Command)
		if !ok {
			return
		}
		ser.processCommand(command)
	case types.LocalEventCommandIndex:
		cIndex, ok := event.Event.(*types.CommandIndex)
		if !ok {
			return
		}
		ser.processCommandIndex(cIndex)
	default:
		ser.logger.Errorf("[%d] Received illegal local event, type: %s", ser.sequencerID, event.Type)
		return
	}
}

func (ser *sequencerImpl) processCommand(command *protos.Command) {
	// Relay the command with pre-defined strategy.
	ser.engine.Sequencing(command)
}

func (ser *sequencerImpl) processCommandIndex(cIndex *types.CommandIndex) {
	if len(ser.commandSet) == 0 {
		ser.timer.StartTimer()
	}

	// command list with receive-order.
	ser.commandSet = append(ser.commandSet, cIndex)
}

func (ser *sequencerImpl) Quit() {
	ser.timer.StopTimer()
	select {
	case <-ser.closeC:
	default:
		close(ser.closeC)
	}
}

func (ser *sequencerImpl) ReceiveLocalEvent(event types.LocalEvent) {
	ser.eventC <- event
}

func (ser *sequencerImpl) generateOrderAttempt() error {
	if len(ser.commandSet) == 0 {
		// skip.
		return nil
	}

	attempt, err := ser.createOrderAttempt()
	if err != nil {
		return fmt.Errorf("create order attempt failed: %s", err)
	}

	// Update the highest order attempt.
	ser.highestAttempt = attempt

	cm, err := protos.PackOrderAttempt(attempt)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	ser.sender.BroadcastPCM(cm)

	ser.logger.Infof("[%d] generated order-attempt %s", ser.sequencerID, attempt.Format())
	return nil
}

func (ser *sequencerImpl) createOrderAttempt() (*protos.OrderAttempt, error) {
	// Sort the command index at first.
	sort.Sort(ser.commandSet)

	// Create content of current order-attempt.
	digestList := make([]string, 0, len(ser.commandSet))
	timestampList := make([]int64, 0, len(ser.commandSet))
	for _, cIndex := range ser.commandSet {
		digestList = append(digestList, cIndex.Digest)
		timestampList = append(timestampList, cIndex.OTime)
	}
	content := protos.NewOrderAttemptContent(digestList, timestampList)
	contentDigest, err := types.CalculateContentDigest(content)
	if err != nil {
		return nil, fmt.Errorf("generate content digest failed: %s", err)
	}

	// Create order attempt.
	attempt := protos.NewOrderAttempt(ser.sequencerID, ser.highestAttempt, contentDigest, content)
	digest, err := types.CalculateOrderAttemptDigest(attempt)
	if err != nil {
		return nil, fmt.Errorf("generate order-attempt digest failed: %s", err)
	}
	attempt.Digest = digest

	// After we have successfully created an order-attempt, reset the command index set.
	ser.commandSet = nil

	return attempt, nil
}
