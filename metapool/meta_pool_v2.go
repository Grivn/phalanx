package metapool

import (
	"github.com/Grivn/phalanx/lib/timer"
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/lib/instance"
	"github.com/Grivn/phalanx/lib/tracker"
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

	// cTracker is used to record the commands received by current node.
	cTracker api.CommandTracker

	// aTracker is used to record the order-attempts received from sequencer.
	aTracker api.AttemptTracker

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
		conf:     conf,
		author:   conf.Author,
		n:        conf.N,
		multi:    conf.Multi,
		quorum:   types.CalculateQuorum(conf.N),
		sequence: uint64(0),
		aggMap:   make(map[string]*protos.PartialOrder),
		replicas: subs,
		pTracker: pTracker,
		cTracker: tracker.NewCommandTracker(conf.Author, conf.Logger),
		aTracker: tracker.NewAttemptTracker(conf.Author, conf.Logger),
		relays:   clients,
		commandC: commandC,
		timer:    timer.NewLocalTimer(conf.Author, timeoutC, conf.Duration, conf.Logger),
		timeoutC: timeoutC,
		closeC:   make(chan bool),
		crypto:   conf.Crypto,
		logger:   conf.Logger,
		metrics:  conf.Metrics,
		commitNo: committedTracker,
		active:   active,
	}
}

//===============================================================
//                 Processor for Local Logs
//===============================================================

func (mp *metaPoolImpl) ProcessCommand(command *protos.Command) {
	// record metrics.
	mp.metrics.ProcessCommand()

	// record the command with command tracker.
	mp.cTracker.RecordCommand(command)
}

func (mp *metaPoolImpl) GenerateProposal() (*protos.Proposal, error) {

	return nil, nil
}

func (mp *metaPoolImpl) fetchCheckpoint(idx types.QueryIndex) *protos.Checkpoint {
	return nil
}

func (mp *metaPoolImpl) VerifyProposal(proposal *protos.Proposal) (types.QueryStream, error) {
	return nil, nil
}

func (mp *metaPoolImpl) ProcessCheckpointRequest(request *protos.CheckpointRequest) {}

func (mp *metaPoolImpl) ProcessCheckpointVote(vote *protos.CheckpointVote) {}

func (mp *metaPoolImpl) ProcessCheckpoint(checkpoint *protos.Checkpoint) {}

//// ProcessVote is used to process the vote message from others.
//// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
//func (mp *metaPoolImpl) ProcessVote(vote *protos.Vote) error {
//	mp.mutex.Lock()
//	defer mp.mutex.Unlock()
//
//	mp.logger.Debugf("[%d] receive vote %s", mp.author, vote.Format())
//
//	// check the existence of order message
//	// here, we should make sure that there is a valid pre-order for us which we have ever assigned.
//	pOrder, ok := mp.aggMap[vote.Digest]
//	if !ok {
//		// there are 2 conditions that a pre-order waiting for agg-sig cannot be found: 1) we have never generated such
//		// a pre-order message, 2) we have already generated an order message for it, which means it has been verified.
//		return nil
//	}
//
//	// verify the signature in vote
//	// here, we would like to check if the signature is valid.
//	if err := mp.crypto.PublicVerify(vote.Certification, types.StringToBytes(vote.Digest), vote.Author); err != nil {
//		return fmt.Errorf("failed to aggregate: %s", err)
//	}
//
//	// record the certification in current vote
//	pOrder.QC.Certs[vote.Author] = vote.Certification
//
//	// check the quorum size for proof-certs
//	if len(pOrder.QC.Certs) == mp.quorum {
//		pOrder.SetOrderedTime()
//
//		mp.logger.Debugf("[%d] found quorum votes, generate quorum order %s", mp.author, pOrder.Format())
//		delete(mp.aggMap, vote.Digest)
//
//		cm, err := protos.PackPartialOrder(pOrder)
//		if err != nil {
//			return fmt.Errorf("generate consensus message error: %s", err)
//		}
//		mp.sender.BroadcastPCM(cm)
//
//		// record metrics.
//		mp.metrics.PartialOrderQuorum(pOrder)
//		return nil
//	}
//
//	mp.logger.Debugf("[%d] aggregate vote for %s, need %d, has %d", mp.author, pOrder.PreOrderDigest(), mp.quorum, len(pOrder.QC.Certs))
//	return nil
//}
//
////===============================================================
////                 Processor for Remote Logs
////===============================================================
//
//// ProcessPreOrder is used to process pre-order messages.
//// We should make sure that we have never received a pre-order/order message
//// whose sequence number is the same as it yet, and we would like to generate a
//// vote message for it if it's legal for us.
//func (mp *metaPoolImpl) ProcessPreOrder(pre *protos.PreOrder) error {
//	return mp.replicas[pre.Author].ReceivePreOrder(pre)
//}
//
//// ProcessPartial is used to process quorum-cert messages.
//// A valid quorum-cert message, which has a series of valid signature which has reached quorum size,
//// could advance the sequence counter. We should record the advanced counter and put the info of
//// order message into the sequential-pool.
//func (mp *metaPoolImpl) ProcessPartial(pOrder *protos.PartialOrder) error {
//	return mp.replicas[pOrder.Author()].ReceivePartial(pOrder)
//}
//
////===============================================================
////                   Read Essential Info
////===============================================================
//
//func (mp *metaPoolImpl) ReadCommand(commandD string) *protos.Command {
//	command := mp.cTracker.ReadCommand(commandD)
//
//	for {
//		if command != nil {
//			break
//		}
//
//		// if we could not read the command, just try the next time.
//		command = mp.cTracker.ReadCommand(commandD)
//	}
//
//	return command
//}
//
//func (mp *metaPoolImpl) ReadPartials(qStream types.QueryStream) []*protos.PartialOrder {
//	var res []*protos.PartialOrder
//
//	for _, qIndex := range qStream {
//		pOrder := mp.pTracker.ReadPartial(qIndex)
//
//		for {
//			if pOrder != nil {
//				break
//			}
//
//			// if we could not read the partial order, just try the next time.
//			pOrder = mp.pTracker.ReadPartial(qIndex)
//		}
//
//		res = append(res, pOrder)
//	}
//
//	return res
//}
//
////=====================================================================
////                  Consensus Proposal Manager
////=====================================================================
//
//func (mp *metaPoolImpl) GenerateProposal() (*protos.PartialOrderBatch, error) {
//	batch := protos.NewPartialOrderBatch(mp.author, mp.n)
//
//	for id, replica := range mp.replicas {
//		index := int(id - 1)
//
//		// read the highest partial order from replica 'id'.
//		hOrder := replica.GetHighOrder()
//
//		if hOrder == nil {
//			// high-order for replica 'id' is nil, record 0 in batch tracker.
//			batch.HighOrders[index] = protos.NewNopPartialOrder()
//			batch.SeqList[index] = 0
//			continue
//		}
//
//		// update batch tracker with information of high-order.
//		batch.HighOrders[index] = hOrder
//		batch.SeqList[index] = hOrder.Sequence()
//	}
//
//	mp.logger.Debugf("[%d] generate batch %s", mp.author, batch.Format())
//	return batch, nil
//}
//
//func (mp *metaPoolImpl) VerifyProposal(batch *protos.PartialOrderBatch) (types.QueryStream, error) {
//	updated := false
//
//	for index, no := range batch.SeqList {
//
//		// calculate the node id.
//		id := uint64(index + 1)
//
//		if no <= mp.commitNo[id] {
//			// committed previous partial order for node id, including partial number 0.
//			mp.logger.Debugf("[%d] haven't updated committed partial order for node %d", mp.author, id)
//			continue
//		}
//
//		pOrder := batch.HighOrders[index]
//
//		if pOrder.Sequence() != no {
//			return nil, fmt.Errorf("invalid partial order seqNo, proposedNo %d, partial seqNo %d", no, pOrder.Sequence())
//		}
//
//		qIndex := types.QueryIndex{Author: pOrder.Author(), SeqNo: pOrder.Sequence()}
//		if !mp.pTracker.IsExist(qIndex) {
//			if err := mp.crypto.VerifyProofCerts(types.StringToBytes(pOrder.PreOrderDigest()), pOrder.QC, mp.quorum); err != nil {
//				return nil, fmt.Errorf("invalid high partial order received from %d: %s", batch.Author, err)
//			}
//		}
//
//		updated = true
//	}
//
//	if !updated {
//		// haven't updated committed partial order.
//		return nil, nil
//	}
//
//	var qStream types.QueryStream
//
//	for index, no := range batch.SeqList {
//		id := uint64(index + 1)
//
//		for {
//			if no <= mp.commitNo[id] {
//				break
//			}
//
//			mp.commitNo[id]++
//
//			qIndex := types.NewQueryIndex(id, mp.commitNo[id])
//			qStream = append(qStream, qIndex)
//		}
//	}
//
//	return qStream, nil
//}
