package memorypool

import (
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/lib/utils"
	"github.com/Grivn/phalanx/metrics"
)

type metaPoolImpl struct {
	conf Config

	//===================================== basic information =========================================

	// mutex is used to deal with the concurrent problems of log-manager.
	mutex sync.RWMutex

	// author is the identifier for current node.
	author uint64

	// n indicates the number of participants in current cluster.
	n int

	// multi indicates the number of proposers each node maintains.
	multi int

	//==================================== sub-chain management =============================================

	// quorum is the legal size for current node.
	quorum int

	// commandSet is used to record the commands' waiting list according to receive order.
	commandSet types.CommandSet

	//===================================== client commands manager ============================================

	// commandC is used to receive the valid transaction from one client instance.
	commandC chan *types.CommandIndex

	// closeC is used to stop log manager.
	closeC chan bool

	//=================================== local timer service ========================================

	// timer is used to control the timeout event to generate order with commands in waiting list.
	timer api.LocalTimer

	// timeoutC is used to receive timeout event.
	timeoutC <-chan bool

	consensusEngine api.ConsensusEngine

	commandTracker api.CommandTracker

	//======================================= consensus manager ============================================

	// commitNo indicates the maximum committed number for each participant's partial order.
	commitNo map[uint64]uint64

	// metrics is used to record the metric info of current node's meta pool.
	metrics *metrics.MetaPoolMetrics
}

func NewMemoryPool(conf Config) *metaPoolImpl {
	conf.Logger.Infof("[%d] initiate log manager, replica count %d", conf.Author, conf.N)

	// initiate communication channel.
	commandC := make(chan *types.CommandIndex)
	timeoutC := make(chan bool)

	// initiate committed number tracker.
	committedTracker := make(map[uint64]uint64)

	return &metaPoolImpl{
		conf:     conf,
		author:   conf.Author,
		n:        conf.N,
		multi:    conf.Multi,
		quorum:   types.CalculateQuorum(conf.N),
		commandC: commandC,
		timer:    utils.NewLocalTimer(conf.Author, timeoutC, conf.Duration, conf.Logger),
		timeoutC: timeoutC,
		closeC:   make(chan bool),
		metrics:  conf.Metrics,
		commitNo: committedTracker,
	}
}

//===============================================================
//                 Processor for Local Logs
//===============================================================

func (mp *metaPoolImpl) ProcessLocalEvent(ev types.LocalEvent) {
	mp.dispatchLocalEvent(ev)
}

func (mp *metaPoolImpl) GenerateProposal() (*protos.Proposal, error) {
	return nil, nil
}

func (mp *metaPoolImpl) VerifyProposal(proposal *protos.Proposal) (types.QueryStream, error) {
	return nil, nil
}

func (mp *metaPoolImpl) dispatchLocalEvent(ev types.LocalEvent) {
	switch ev.Type {
	case types.LocalEventCommand:
		command, ok := ev.Event.(*protos.Command)
		if !ok {
			return
		}
		go mp.processCommand(command)
	default:
		return
	}
}

func (mp *metaPoolImpl) processCommand(command *protos.Command) {
	// record metrics.
	mp.metrics.ProcessCommand()

	// record the command with command tracker.
	mp.commandTracker.Record(command)
}

func (mp *metaPoolImpl) processConsensusMessage(message *protos.ConsensusMessage) {
	mp.consensusEngine.ProcessConsensusMessage(message)
}
