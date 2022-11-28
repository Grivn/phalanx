package sequencer

import (
	"fmt"
	"sort"
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/lib/instance"
	"github.com/Grivn/phalanx/lib/utils"
	"github.com/Grivn/phalanx/metrics"
)

type sequencerImpl struct {
	//===================================== basic information =========================================

	// mutex is used to deal with the concurrent problems of log-manager.
	mutex sync.RWMutex

	// author is the identifier for current node.
	author uint64

	// byz indicates if current node is the adversary.
	byz bool

	// snapping indicates if we have started the situation for snapping up.
	snapping bool

	first bool

	//==================================== sub-chain management =============================================

	// sequence is a target for local-log.
	sequence uint64

	//
	highAttempt *protos.OrderAttempt

	// commandSet is used to record the commands' waiting list according to receive order.
	commandSet types.CommandSet

	//===================================== client commands manager ============================================

	// clients are used to track the commands send from them.
	clients map[uint64]api.ClientInstance

	// active indicates the number of active client instance.
	active *int64

	// transactionC is used to receive transactions.
	transactionC chan *protos.Transaction

	// commandC is used to receive the valid transaction from one client instance.
	commandC chan *types.CommandIndex

	// closeC is used to stop log manager.
	closeC chan bool

	//=================================== local timer service ========================================

	// timer is used to control the timeout event to generate order with commands in waiting list.
	timer api.LocalTimer

	// timeoutC is used to receive timeout event.
	timeoutC <-chan bool

	//======================================= consensus manager ============================================

	// commitNo indicates the maximum committed number for each participant's partial order.
	commitNo map[uint64]uint64

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
	commandC := make(chan *types.CommandIndex, 100)
	timeoutC := make(chan bool)

	// initiate committed number tracker.
	committedTracker := make(map[uint64]uint64)

	// initiate active client count pointer.
	active := new(int64)
	*active = int64(0)

	// initiate client instances.
	clients := make(map[uint64]api.ClientInstance)
	for i := 0; i < conf.N*conf.Multi; i++ {
		id := uint64(i + 1)
		client := instance.NewRelay(conf.Author, id, commandC, active, conf.Logger)
		clients[id] = client
	}

	return &sequencerImpl{
		author:   conf.Author,
		sequence: uint64(0),
		clients:  clients,
		commandC: commandC,
		timer:    utils.NewLocalTimer(conf.Author, timeoutC, conf.Duration, conf.Logger),
		timeoutC: timeoutC,
		closeC:   make(chan bool),
		sender:   conf.Sender,
		logger:   conf.Logger,
		metrics:  conf.Metrics,
		commitNo: committedTracker,
		active:   active,
		byz:      conf.Byz,
	}
}

func (ser *sequencerImpl) Run() {
	for {
		select {
		case <-ser.closeC:
			return
		case c := <-ser.commandC:
			ser.appendCommandIndex(c)
		case <-ser.timeoutC:
			if err := ser.tryGenerateOrderAttempt(); err != nil {
				panic(fmt.Sprintf("log manager runtime error: %s", err))
			}
		}
	}
}

func (ser *sequencerImpl) Quit() {
	ser.timer.StopTimer()
	select {
	case <-ser.closeC:
	default:
		close(ser.closeC)
	}
}

func (ser *sequencerImpl) ReceiveTxs(transaction *protos.Transaction) {

}

func (ser *sequencerImpl) ProcessCommand(command *protos.Command) {
	if ser.byz && ser.snapping && command.Author != ser.author {
		// current node is the arbitrary
		// it is in snapping up situation.
		return
	}

	// relay the command with pre-defined strategy.
	ser.clientInstanceReminder(command)
}

func (ser *sequencerImpl) clientInstanceReminder(command *protos.Command) {
	// select the client.
	client, ok := ser.clients[command.Author]
	if !ok {
		// if there is not a client instance, initiate it.
		// NOTE: concurrency problem.
		ser.logger.Errorf("[%d] don't have client instance %d, initiate it", ser.author, command.Author)
		client = instance.NewClient(ser.author, command.Author, ser.commandC, ser.active, ser.logger)
		ser.clients[command.Author] = client
	}

	// append the transaction into this client.
	client.Append(command)
}

func (ser *sequencerImpl) checkHighOrder() error {
	// here, we should make sure the highest sequence number is valid.
	if ser.highAttempt == nil {
		switch ser.sequence {
		case 0:
			// if there isn't any high order, we should make sure that we are trying to generate the first partial order.
			return nil
		default:
			return fmt.Errorf("invalid status for current node, highest order nil, current seqNo %d", ser.sequence)
		}
	}

	if ser.highAttempt.SeqNo != ser.sequence {
		return fmt.Errorf("invalid status for current node, highest order %d, current seqNo %d", ser.highAttempt.SeqNo, ser.sequence)
	}

	// highest partial order has a valid sequence number.
	return nil
}

func (ser *sequencerImpl) updateHighAttempt(attempt *protos.OrderAttempt) {
	ser.highAttempt = attempt
}

// appendCommandIndex is used to append the received command index into the command set.
func (ser *sequencerImpl) appendCommandIndex(cIndex *types.CommandIndex) {
	ser.mutex.Lock()
	defer ser.mutex.Unlock()

	if len(ser.commandSet) == 0 {
		ser.timer.StartTimer()
	}

	// command list with receive-order.
	ser.commandSet = append(ser.commandSet, cIndex)
}

func (ser *sequencerImpl) tryGenerateOrderAttempt() error {
	ser.mutex.Lock()
	defer ser.mutex.Unlock()

	// timeout event generate order.
	ser.logger.Debugf("[%d] order attempt generation timer expired", ser.author)
	return ser.generateOrderAttempt()
}

func (ser *sequencerImpl) generateOrderAttempt() error {
	if len(ser.commandSet) == 0 {
		// skip.
		return nil
	}

	// make sure the highest partial order has a valid status.
	if err := ser.checkHighOrder(); err != nil {
		return fmt.Errorf("highest partial order error: %s", err)
	}

	// advance the sequence number.
	ser.sequence++

	digestList := make([]string, len(ser.commandSet))
	timestampList := make([]int64, len(ser.commandSet))

	sort.Sort(ser.commandSet)
	if ser.byz && !ser.snapping {
		// current node is the arbitrary, and it's not snapping up situation.
		timeSet := make([]int64, len(ser.commandSet))
		byz := make(types.ByzCommandSet, len(ser.commandSet))
		for index, command := range ser.commandSet {
			timeSet[index] = command.OTime
		}
		for index, command := range ser.commandSet {
			byz[index] = command
		}
		sort.Sort(byz)
		ser.commandSet = nil
		for index, command := range byz {
			command.OTime = timeSet[index]
			ser.commandSet = append(ser.commandSet, command)
		}
	}
	for i, cIndex := range ser.commandSet {
		digestList[i] = cIndex.Digest
		timestampList[i] = cIndex.OTime

		// record metrics.
		ser.metrics.SelectCommand(cIndex)
	}

	// generate order-attempt message.

	content := protos.NewOrderAttemptContent(digestList, timestampList)

	contentDigest, err := types.CalculateContentDigest(content)
	if err != nil {
		return fmt.Errorf("generate content digest failed: %s", err)
	}

	attempt := protos.NewOrderAttempt(ser.author, ser.sequence, ser.highAttempt, contentDigest, content)

	digest, err := types.CalculateOrderAttemptDigest(attempt)
	if err != nil {
		return fmt.Errorf("generate order-attempt digest failed: %s", err)
	}

	attempt.Digest = digest

	// reset receive-order lists.
	ser.commandSet = nil

	ser.logger.Infof("[%d] generate order-attempt %s", ser.author, attempt.Format())
	ser.updateHighAttempt(attempt)

	cm, err := protos.PackOrderAttempt(attempt)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	ser.sender.BroadcastPCM(cm)

	// record metrics.
	ser.metrics.GenerateOrder()
	return nil
}
