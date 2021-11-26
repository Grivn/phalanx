package metapool

import (
	"fmt"
	"sync"
	"time"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/metapool/instance"
	"github.com/Grivn/phalanx/metapool/tracker"
)

// todo pre-trusted order entry
// while receiving a quorum cert in cross-consensus:
// 1) qc with lower or equal height: skip
// 2) qc with higher height: update the trusted number and generate QC cert for it
//
// while we receive a pre-order, check if there is already QC for it:
// 1) there isn't QC cert: normal process
// 2) there is a QC cert: check the digest between them, if they are not equal, reject.

type metaPool struct {
	//===================================== basic information =========================================

	// mutex is used to deal with the concurrent problems of log-manager.
	mutex sync.Mutex

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

	// highOrder is the highest partial order for current chained manager.
	// as for that we should be responsible for the order of our own private chain, each block could
	highOrder *protos.PreOrder

	// aggMap is used to generate aggregated-certificates.
	aggMap map[string]*protos.PartialOrder

	// replicas is the module for us to process consensus messages for participates.
	// when we try to read the partial order to execute, we should read them from each sub instance.
	replicas map[uint64]internal.ReplicaInstance

	// pTracker is used to record the partial orders received by current node.
	pTracker internal.PartialTracker

	// commandList is used to record the commands' waiting list according to receive order.
	commandList []string

	// timestampList is used to record the timestamps when the command has been selected into waiting list.
	timestampList []int64

	//===================================== client commands manager ============================================

	// cTracker is used to record the commands received by current node.
	cTracker internal.CommandTracker

	// clients are used to track the commands send from them.
	clients map[uint64]internal.ClientInstance

	// isActive indicates the active client instance which has pending commands to be committed.
	isActive map[uint64]bool

	// commandC is used to receive the valid transaction from one client instance.
	commandC chan *types.CommandIndex

	// closeC is used to stop log manager.
	closeC chan bool

	//=================================== local timer service ========================================

	// timer is used to control the timeout event to generate order with commands in waiting list.
	timer *localTimer

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
}

func NewMetaPool(n, multi int, author uint64, sender external.NetworkService, logger external.Logger) internal.MetaPool {
	logger.Infof("[%d] initiate log manager, replica count %d", author, n)

	// initiate committed number tracker.
	committedTracker := make(map[uint64]uint64)

	// initiate a partial tracker for current node.
	pTracker := tracker.NewPartialTracker(author, logger)

	subs := make(map[uint64]internal.ReplicaInstance)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		subs[id] = instance.NewReplicaInstance(author, id, pTracker, sender, logger)
		committedTracker[id] = 0
	}

	timeoutC := make(chan bool)

	return &metaPool{
		author:   author,
		n:        n,
		multi:    multi,
		quorum:   types.CalculateQuorum(n),
		sequence: uint64(0),
		aggMap:   make(map[string]*protos.PartialOrder),
		replicas: subs,
		pTracker: pTracker,
		cTracker: tracker.NewCommandTracker(author, logger),
		clients:  make(map[uint64]internal.ClientInstance),
		isActive: make(map[uint64]bool),
		commandC: make(chan *types.CommandIndex, 10000),
		timer:    newLocalTimer(timeoutC, 500*time.Millisecond),
		timeoutC: timeoutC,
		closeC:   make(chan bool),
		sender:   sender,
		logger:   logger,
		commitNo: committedTracker,
	}
}

func (mp *metaPool) Run() {
	for {
		select {
		case <-mp.closeC:
			return
		case c := <-mp.commandC:
			if err := mp.tryGeneratePreOrder(c); err != nil {
				panic(fmt.Sprintf("log manager runtime error: %s", err))
			}
		case <-mp.timeoutC:
			if err := mp.tryGeneratePreOrder(nil); err != nil {
				panic(fmt.Sprintf("log manager runtime error: %s", err))
			}
		}
	}
}

func (mp *metaPool) Quit() {
	mp.timer.stopTimer()
	select {
	case <-mp.closeC:
	default:
		close(mp.closeC)
	}
}

func (mp *metaPool) Committed(author uint64, seqNo uint64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	remain := mp.clients[author].Commit(seqNo)

	if remain == 0 {
		delete(mp.isActive, author)
	}
}

//===============================================================
//                 Processor for Local Logs
//===============================================================

func (mp *metaPool) ProcessCommand(command *protos.Command) {
	// record the command with command tracker.
	mp.cTracker.RecordCommand(command)

	// select the client instance and record the command target.
	mp.clientInstanceReminder(command)
}

func (mp *metaPool) clientInstanceReminder(command *protos.Command) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// select the client.
	client, ok := mp.clients[command.Author]
	if !ok {
		// if there is not a client instance, initiate it.
		client = instance.NewClient(mp.author, command.Author, mp.commandC, mp.logger)
		mp.clients[command.Author] = client
	}
	mp.isActive[command.Author] = true

	// append the transaction into this client.
	remain := client.Append(command)

	if remain == 0 {
		delete(mp.isActive, command.Author)
	}
}

func (mp *metaPool) checkHighOrder() error {

	// here, we should make sure the highest sequence number is valid.
	if mp.highOrder == nil {
		switch mp.sequence {
		case 0:
			// if there isn't any high order, we should make sure that we are trying to generate the first partial order.
			return nil
		default:
			return fmt.Errorf("invalid status for current node, highest order nil, current seqNo %d", mp.sequence)
		}
	}

	if mp.highOrder.Sequence != mp.sequence {
		return fmt.Errorf("invalid status for current node, highest order %d, current seqNo %d", mp.highOrder.Sequence, mp.sequence)
	}

	// highest partial order has a valid sequence number.
	return nil
}

func (mp *metaPool) updateHighOrder(pre *protos.PreOrder) {
	mp.highOrder = pre
}

// tryGeneratePreOrder is used to process the command received from one client instance.
// We would like to assign the latest seqNo for it and generate a pre-order message.
func (mp *metaPool) tryGeneratePreOrder(cIndex *types.CommandIndex) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if cIndex == nil {
		// timeout event generate order.
		return mp.generateOrder()
	}

	if len(mp.commandList) == 0 {
		mp.timer.startTimer()
	}

	// command list with receive-order.
	mp.commandList = append(mp.commandList, cIndex.Digest)
	mp.timestampList = append(mp.timestampList, time.Now().UnixNano())

	if len(mp.commandList) < len(mp.isActive) {
		// skip
		return nil
	}
	mp.timer.stopTimer()
	return mp.generateOrder()
}

func (mp *metaPool) generateOrder() error {

	// make sure the highest partial order has a valid status.
	if err := mp.checkHighOrder(); err != nil {
		return fmt.Errorf("highest partial order error: %s", err)
	}

	// advance the sequence number.
	mp.sequence++

	// generate pre order message.
	pre := protos.NewPreOrder(mp.author, mp.sequence, mp.commandList, mp.timestampList, mp.highOrder)
	digest, err := crypto.CalculateDigest(pre)
	if err != nil {
		return fmt.Errorf("pre order marshal error: %s", err)
	}
	pre.Digest = digest

	// reset receive-order lists.
	mp.commandList = nil
	mp.timestampList = nil

	// generate self-signature for current pre-order
	signature, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(mp.author))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}

	// init the order message in aggregate map and assign self signature
	mp.aggMap[pre.Digest] = protos.NewPartialOrder(pre)
	mp.aggMap[pre.Digest].QC.Certs[mp.author] = signature

	mp.logger.Infof("[%d] generate pre-order %s", mp.author, pre.Format())

	// update the highest pre order for current node.
	mp.updateHighOrder(pre)

	cm, err := protos.PackPreOrder(pre)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	mp.sender.BroadcastPCM(cm)
	return nil
}

// ProcessVote is used to process the vote message from others.
// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
func (mp *metaPool) ProcessVote(vote *protos.Vote) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	mp.logger.Debugf("[%d] receive vote %s", mp.author, vote.Format())

	// check the existence of order message
	// here, we should make sure that there is a valid pre-order for us which we have ever assigned.
	pOrder, ok := mp.aggMap[vote.Digest]
	if !ok {
		// there are 2 conditions that a pre-order waiting for agg-sig cannot be found: 1) we have never generated such
		// a pre-order message, 2) we have already generated an order message for it, which means it has been verified.
		return nil
	}

	// verify the signature in vote
	// here, we would like to check if the signature is valid.
	if err := crypto.PubVerify(vote.Certification, types.StringToBytes(vote.Digest), int(vote.Author)); err != nil {
		return fmt.Errorf("failed to aggregate: %s", err)
	}

	// record the certification in current vote
	pOrder.QC.Certs[vote.Author] = vote.Certification

	// check the quorum size for proof-certs
	if len(pOrder.QC.Certs) == mp.quorum {
		mp.logger.Debugf("[%d] found quorum votes for pre-order %s, generate quorum order", mp.author, pOrder.PreOrderDigest())
		delete(mp.aggMap, vote.Digest)

		cm, err := protos.PackPartialOrder(pOrder)
		if err != nil {
			return fmt.Errorf("generate consensus message error: %s", err)
		}
		mp.sender.BroadcastPCM(cm)
		return nil
	}

	mp.logger.Debugf("[%d] aggregate vote for %s, need %d, has %d", mp.author, pOrder.PreOrderDigest(), mp.quorum, len(pOrder.QC.Certs))
	return nil
}

//===============================================================
//                 Processor for Remote Logs
//===============================================================

// ProcessPreOrder is used to process pre-order messages.
// We should make sure that we have never received a pre-order/order message
// whose sequence number is the same as it yet, and we would like to generate a
// vote message for it if it's legal for us.
func (mp *metaPool) ProcessPreOrder(pre *protos.PreOrder) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	return mp.replicas[pre.Author].ReceivePreOrder(pre)
}

// ProcessPartial is used to process quorum-cert messages.
// A valid quorum-cert message, which has a series of valid signature which has reached quorum size,
// could advance the sequence counter. We should record the advanced counter and put the info of
// order message into the sequential-pool.
func (mp *metaPool) ProcessPartial(pOrder *protos.PartialOrder) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	return mp.replicas[pOrder.Author()].ReceivePartial(pOrder)
}

//===============================================================
//                   Read Essential Info
//===============================================================

func (mp *metaPool) ReadCommand(commandD string) *protos.Command {
	command := mp.cTracker.ReadCommand(commandD)

	for {
		if command != nil {
			break
		}

		// if we could not read the command, just try the next time.
		command = mp.cTracker.ReadCommand(commandD)
	}

	return command
}

func (mp *metaPool) ReadPartials(qStream types.QueryStream) []*protos.PartialOrder {
	var res []*protos.PartialOrder

	for _, qIndex := range qStream {
		pOrder := mp.pTracker.ReadPartial(qIndex)

		for {
			if pOrder != nil {
				break
			}

			// if we could not read the partial order, just try the next time.
			pOrder = mp.pTracker.ReadPartial(qIndex)
		}

		res = append(res, pOrder)
	}

	return res
}

//=====================================================================
//                  Consensus Proposal Manager
//=====================================================================

func (mp *metaPool) GenerateProposal() (*protos.PartialOrderBatch, error) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	batch := protos.NewPartialOrderBatch(mp.author, mp.n)

	for id, replica := range mp.replicas {
		index := int(id-1)

		// read the highest partial order from replica 'id'.
		hOrder := replica.GetHighOrder()

		if hOrder == nil {
			// high-order for replica 'id' is nil, record 0 in batch tracker.
			batch.HighOrders[index] = protos.NewNopPartialOrder()
			batch.SeqList[index] = 0
			continue
		}

		// update batch tracker with information of high-order.
		batch.HighOrders[index] = hOrder
		batch.SeqList[index] = hOrder.Sequence()
	}

	mp.logger.Debugf("[%d] generate batch %s", mp.author, batch.Format())
	return batch, nil
}

func (mp *metaPool) VerifyProposal(batch *protos.PartialOrderBatch) (types.QueryStream, error) {
	updated := false

	for index, no := range batch.SeqList {

		// calculate the node id.
		id := uint64(index+1)

		if no <= mp.commitNo[id] {
			// committed previous partial order for node id, including partial number 0.
			mp.logger.Debugf("[%d] haven't updated committed partial order for node %d", mp.author, id)
			continue
		}

		pOrder := batch.HighOrders[index]

		if pOrder.Sequence() != no {
			return nil, fmt.Errorf("invalid partial order seqNo, proposedNo %d, partial seqNo %d", no, pOrder.Sequence())
		}

		qIndex := types.QueryIndex{Author: pOrder.Author(), SeqNo: pOrder.Sequence()}
		if !mp.pTracker.IsExist(qIndex) {
			if err := crypto.VerifyProofCerts(types.StringToBytes(pOrder.PreOrderDigest()), pOrder.QC, mp.quorum); err != nil {
				return nil, fmt.Errorf("invalid high partial order received from %d: %s", batch.Author, err)
			}
		}

		updated = true
	}

	if !updated {
		// haven't updated committed partial order.
		return nil, nil
	}

	var qStream types.QueryStream

	for index, no := range batch.SeqList {
		id := uint64(index+1)

		for {
			if no <= mp.commitNo[id] {
				break
			}

			mp.commitNo[id]++

			qIndex := types.NewQueryIndex(id, mp.commitNo[id])
			qStream = append(qStream, qIndex)
		}
	}

	return qStream, nil
}
