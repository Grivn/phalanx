package sequencer

import (
	"fmt"
	"sort"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/service/seqengine"
	"github.com/Grivn/phalanx/pkg/utils/timer"
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

	// closeC is used to stop log manager.
	closeC chan bool

	//=================================== local timer service ========================================

	// singleTimer is used to control the timeout event to generate order with commands in waiting list.
	singleTimer api.SingleTimer

	// engine is used to sequencing the commands received and create command_index to propose order-attempts.
	engine api.SequencingEngine

	//======================================= external tools ===========================================

	// sender is used to send consensus message into network.
	sender external.Sender

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencer(conf Config) *sequencerImpl {
	conf.Logger.Infof("[%d] initiate log manager, replica count %d", conf.Author, conf.N)
	return &sequencerImpl{
		sequencerID: conf.Author,
		singleTimer: timer.NewSingleTimer(conf.Duration, conf.Logger),
		closeC:      make(chan bool),
		engine:      seqengine.NewSequencingEngine(conf.Author, conf.Logger),
		sender:      conf.Sender,
		logger:      conf.Logger,
		byz:         conf.Byz,
	}
}

func (ser *sequencerImpl) Run() {
	go ser.listener()
}

func (ser *sequencerImpl) Quit() {
	ser.singleTimer.StopTimer()
	select {
	case <-ser.closeC:
	default:
		close(ser.closeC)
	}
}

func (ser *sequencerImpl) Sequencing(command *protos.Command) {
	ser.processCommand(command)
}

func (ser *sequencerImpl) listener() {
	for {
		select {
		case <-ser.closeC:
			return
		case cIndex := <-ser.engine.CommandIndexChan():
			ser.processCommandIndex(cIndex)
		case <-ser.singleTimer.TimeoutChan():
			if err := ser.generateOrderAttempt(); err != nil {
				panic(fmt.Sprintf("log manager runtime error: %s", err))
			}
		}
	}
}

func (ser *sequencerImpl) processCommand(command *protos.Command) {
	ser.engine.Sequencing(command)
}

func (ser *sequencerImpl) processCommandIndex(cIndex *types.CommandIndex) {
	if len(ser.commandSet) == 0 {
		ser.singleTimer.StartTimer()
	}

	// command list with receive-order.
	ser.commandSet = append(ser.commandSet, cIndex)
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
