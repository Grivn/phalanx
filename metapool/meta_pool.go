package metapool

import (
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/event"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/lib/instance"
	"github.com/Grivn/phalanx/lib/tracker"
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

	// sequence is a target for local-log.
	sequence uint64

	//
	highAttempt *protos.OrderAttempt

	// aggMap is used to generate aggregated-certificates.
	aggMap map[string]*protos.PartialOrder

	// replicas is the module for us to process consensus messages for participates.
	// when we try to read the partial order to execute, we should read them from each sub instance.
	replicas map[uint64]api.ReplicaInstance

	// pTracker is used to record the partial orders received by current node.
	pTracker api.PartialTracker

	// commandSet is used to record the commands' waiting list according to receive order.
	commandSet types.CommandSet

	//===================================== client commands manager ============================================

	// commandTracker is used to record the commands received by current node.
	commandTracker api.CommandTracker

	// attemptTracker is used to record the order-attempts received from sequencer.
	attemptTracker api.AttemptTracker

	// checkpointTracker is used record the checkpoints received from consensus node.
	checkpointTracker api.CheckpointTracker

	// are used to track the commands send from them.
	relays map[uint64]api.Relay

	// active indicates the number of active client instance.
	active *int64

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

	// consensusEngine is used to generate checkpoint for order-attempts.
	consensusEngine api.ConsensusEngine

	//==================================== crypto management =============================================

	// crypto is used to generate/verify certificates.
	crypto api.Crypto

	//======================================= external tools ===========================================

	// logger is used to print logs.
	logger external.Logger

	// metrics is used to record the metric info of current node's meta pool.
	metrics *metrics.MetaPoolMetrics
}

func NewMetaPool(conf Config) *metaPoolImpl {
	conf.Logger.Infof("[%d] initiate log manager, replica count %d", conf.Author, conf.N)

	// initiate communication channel.
	commandC := make(chan *types.CommandIndex, 100)
	timeoutC := make(chan bool)

	// initiate committed number tracker.
	committedTracker := make(map[uint64]uint64)

	// initiate a partial tracker for current node.
	pTracker := tracker.NewPartialTracker(conf.Author, conf.Logger)

	// initiate replica instances.
	subs := make(map[uint64]api.ReplicaInstance)
	for i := 0; i < conf.N; i++ {
		id := uint64(i + 1)
		subs[id] = instance.NewReplicaInstance(conf.Author, id, pTracker, conf.Crypto, conf.Sender, conf.Logger)
		committedTracker[id] = 0
	}

	// initiate active client count pointer.
	active := new(int64)
	*active = int64(0)

	// initiate client instances.
	clients := make(map[uint64]api.Relay)
	for i := 0; i < conf.N*conf.Multi; i++ {
		id := uint64(i + 1)
		client := instance.NewRelay(conf.Author, id, commandC, active, conf.Logger)
		clients[id] = client
	}

	return &metaPoolImpl{
		conf:           conf,
		author:         conf.Author,
		n:              conf.N,
		multi:          conf.Multi,
		quorum:         types.CalculateQuorum(conf.N),
		sequence:       uint64(0),
		aggMap:         make(map[string]*protos.PartialOrder),
		replicas:       subs,
		pTracker:       pTracker,
		commandTracker: tracker.NewCommandTracker(conf.Author, conf.Logger),
		attemptTracker: tracker.NewAttemptTracker(conf.Author, conf.Logger),
		relays:         clients,
		commandC:       commandC,
		timer:          utils.NewLocalTimer(conf.Author, timeoutC, conf.Duration, conf.Logger),
		timeoutC:       timeoutC,
		closeC:         make(chan bool),
		crypto:         conf.Crypto,
		logger:         conf.Logger,
		metrics:        conf.Metrics,
		commitNo:       committedTracker,
		active:         active,
	}
}

//===============================================================
//                 Processor for Local Logs
//===============================================================

func (mp *metaPoolImpl) ProcessLocalEvent(ev event.LocalEvent) {
	mp.dispatchLocalEvent(ev)
}

func (mp *metaPoolImpl) GenerateProposal() (*protos.Proposal, error) {
	return nil, nil
}

func (mp *metaPoolImpl) VerifyProposal(proposal *protos.Proposal) (types.QueryStream, error) {
	return nil, nil
}

func (mp *metaPoolImpl) dispatchLocalEvent(ev event.LocalEvent) {
	switch ev.Type {
	case event.LocalEventCommand:
		command, ok := ev.Event.(*protos.Command)
		if !ok {
			return
		}
		go mp.processCommand(command)
	case event.LocalEventConsensusMessage:
		message, ok := ev.Event.(*protos.ConsensusMessage)
		if !ok {
			return
		}
		go mp.processConsensusMessage(message)
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
