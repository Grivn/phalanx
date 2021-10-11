package logmanager

import (
	"fmt"
	"sync"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

// todo pre-trusted order entry
// while receiving a quorum cert in cross-consensus:
// 1) qc with lower or equal height: skip
// 2) qc with higher height: update the trusted number and generate QC cert for it
//
// while we receive a pre-order, check if there is already QC for it:
// 1) there isn't QC cert: normal process
// 2) there is a QC cert: check the digest between them, if they are not equal, reject.

type logManager struct {
	//===================================== basic information =========================================

	// mutex is used to deal with the concurrent problems of log-manager.
	mutex sync.Mutex

	// author is the identifier for current node.
	author uint64

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

	// subs is the module for us to process consensus messages for participates.
	// when we try to read the partial order to execute, we should read them from each sub instance.
	subs map[uint64]*subInstance

	// pTracker is used to record the partial orders received by current node.
	pTracker *partialTracker

	//===================================== client commands manager ============================================

	// cTracker is used to record the commands received by current node.
	cTracker *commandTracker

	// clients are used to track the commands send from them.
	clients map[uint64]*clientInstance

	// commandC is used to receive the valid transaction from one client instance.
	commandC chan *protos.Command

	// closeC is used to stop log manager.
	closeC chan bool

	//======================================= consensus manager ============================================

	//
	commitNo map[uint64]uint64

	//======================================= external tools ===========================================

	// sender is used to send consensus message into network.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func NewLogManager(n int, author uint64, sender external.NetworkService, logger external.Logger) *logManager {
	logger.Infof("[%d] initiate log manager, replica count %d", author, n)

	// initiate committed number tracker.
	committedTracker := make(map[uint64]uint64)

	// initiate a partial tracker for current node.
	pTracker := newPartialTracker(author, logger)

	subs := make(map[uint64]*subInstance)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		subs[id] = newSubInstance(author, id, pTracker, sender, logger)
		committedTracker[id] = 0
	}

	return &logManager{
		quorum:   types.CalculateQuorum(n),
		author:   author,
		sequence: uint64(0),
		aggMap:   make(map[string]*protos.PartialOrder),
		subs:     subs,
		pTracker: pTracker,
		cTracker: newCommandTracker(author, logger),
		clients:  make(map[uint64]*clientInstance),
		commandC: make(chan *protos.Command),
		closeC:   make(chan bool),
		sender:   sender,
		logger:   logger,
		commitNo: committedTracker,
	}
}

func (mgr *logManager) Run() {
	for {
		select {
		case <-mgr.closeC:
			return
		case c := <-mgr.commandC:
			if err := mgr.tryGeneratePreOrder(c); err != nil {
				panic(fmt.Sprintf("log manager runtime error: %s", err))
			}
		}
	}
}

func (mgr *logManager) Committed(author uint64, seqNo uint64) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	go mgr.clients[author].commit(seqNo)
}

//===============================================================
//                 Processor for Local Logs
//===============================================================

func (mgr *logManager) ProcessCommand(command *protos.Command) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	// record the command with command tracker.
	mgr.cTracker.recordCommand(command)

	// select the client instance according to the identifier of command author.
	client, ok := mgr.clients[command.Author]
	if !ok {
		// if we cannot find the client, initiate an instance for this client.
		client = newClient(mgr.author, command.Author, mgr.commandC, mgr.logger)
		mgr.clients[command.Author] = client
	}

	// append the transaction into this client.
	client.append(command)

	return nil
}

func (mgr *logManager) checkHighOrder() error {

	// here, we should make sure the highest sequence number is valid.
	if mgr.highOrder == nil {
		switch mgr.sequence {
		case 0:
			// if there isn't any high order, we should make sure that we are trying to generate the first partial order.
			return nil
		default:
			return fmt.Errorf("invalid status for current node, highest order nil, current seqNo %d", mgr.sequence)
		}
	}

	if mgr.highOrder.Sequence != mgr.sequence {
		return fmt.Errorf("invalid status for current node, highest order %d, current seqNo %d", mgr.highOrder.Sequence, mgr.sequence)
	}

	// highest partial order has a valid sequence number.
	return nil
}

func (mgr *logManager) updateHighOrder(pre *protos.PreOrder) {
	mgr.highOrder = pre
}

// tryGeneratePreOrder is used to process the command received from one client instance.
// We would like to assign the latest seqNo for it and generate a pre-order message.
func (mgr *logManager) tryGeneratePreOrder(command *protos.Command) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	// make sure the highest partial order has a valid status.
	if err := mgr.checkHighOrder(); err != nil {
		return fmt.Errorf("highest partial order error: %s", err)
	}

	// advance the sequence number.
	mgr.sequence++

	// generate pre order message.
	pre := protos.NewPreOrder(mgr.author, mgr.sequence, command, mgr.highOrder)
	digest, err := crypto.CalculateDigest(pre)
	if err != nil {
		return fmt.Errorf("pre order marshal error: %s", err)
	}
	pre.Digest = digest

	// generate self-signature for current pre-order
	signature, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(mgr.author))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}

	// init the order message in aggregate map and assign self signature
	mgr.aggMap[pre.Digest] = protos.NewPartialOrder(pre)
	mgr.aggMap[pre.Digest].QC.Certs[mgr.author] = signature

	mgr.logger.Infof("[%d] generate pre-order %s", mgr.author, pre.Format())

	// update the highest pre order for current node.
	mgr.updateHighOrder(pre)

	cm, err := protos.PackPreOrder(pre)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	mgr.sender.BroadcastPCM(cm)
	return nil
}

// ProcessVote is used to process the vote message from others.
// It could aggregate a agg-signature for one pre-order and generate an order message for one command.
func (mgr *logManager) ProcessVote(vote *protos.Vote) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	mgr.logger.Debugf("[%d] receive vote %s", mgr.author, vote.Format())

	// check the existence of order message
	// here, we should make sure that there is a valid pre-order for us which we have ever assigned.
	pOrder, ok := mgr.aggMap[vote.Digest]
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
	if len(pOrder.QC.Certs) == mgr.quorum {
		mgr.logger.Debugf("[%d] found quorum votes for pre-order %s, generate quorum order", mgr.author, pOrder.PreOrderDigest())
		delete(mgr.aggMap, vote.Digest)

		cm, err := protos.PackPartialOrder(pOrder)
		if err != nil {
			return fmt.Errorf("generate consensus message error: %s", err)
		}
		mgr.sender.BroadcastPCM(cm)
		return nil
	}

	mgr.logger.Debugf("[%d] aggregate vote for %s, need %d, has %d", mgr.author, pOrder.PreOrderDigest(), mgr.quorum, len(pOrder.QC.Certs))
	return nil
}

//===============================================================
//                 Processor for Remote Logs
//===============================================================

// ProcessPreOrder is used to process pre-order messages.
// We should make sure that we have never received a pre-order/order message
// whose sequence number is the same as it yet, and we would like to generate a
// vote message for it if it's legal for us.
func (mgr *logManager) ProcessPreOrder(pre *protos.PreOrder) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	return mgr.subs[pre.Author].processPreOrder(pre)
}

// ProcessPartial is used to process quorum-cert messages.
// A valid quorum-cert message, which has a series of valid signature which has reached quorum size,
// could advance the sequence counter. We should record the advanced counter and put the info of
// order message into the sequential-pool.
func (mgr *logManager) ProcessPartial(pOrder *protos.PartialOrder) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	return mgr.subs[pOrder.Author()].processPartial(pOrder)
}

//===============================================================
//                   Read Essential Info
//===============================================================

func (mgr *logManager) ReadCommand(commandD string) *protos.Command {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	command := mgr.cTracker.readCommand(commandD)

	for {
		if command != nil {
			break
		}

		// if we could not read the command, just try the next time.
		command = mgr.cTracker.readCommand(commandD)
	}

	return command
}

func (mgr *logManager) ReadPartials(qStream types.QueryStream) []*protos.PartialOrder {
	var res []*protos.PartialOrder

	for _, qIndex := range qStream {
		pOrder := mgr.pTracker.readPartial(qIndex)

		for {
			if pOrder != nil {
				break
			}

			// if we could not read the partial order, just try the next time.
			pOrder = mgr.pTracker.readPartial(qIndex)
		}

		res = append(res, pOrder)
	}

	return res
}

//=====================================================================
//                  Consensus Proposal Manager
//=====================================================================

func (mgr *logManager) GenerateProposal() (*protos.PartialOrderBatch, error) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	batch := protos.NewPartialOrderBatch(mgr.author)

	for _, sub := range mgr.subs {
		if sub.highPartialOrder == nil {
			mgr.logger.Debugf("[%d] generate batch, nil high order for node %d", mgr.author, sub.id)
			batch.ProposedNos[sub.id] = 0
			continue
		}

		batch.Partials[sub.id] = sub.highPartialOrder
		batch.ProposedNos[sub.id] = sub.highPartialOrder.Sequence()
		mgr.logger.Debugf("[%d] generate batch, high order for node %d, %s", mgr.author, sub.id, sub.highPartialOrder.Format())
	}

	return batch, nil
}

func (mgr *logManager) VerifyProposal(batch *protos.PartialOrderBatch) (types.QueryStream, error) {
	updated := false

	for id, no := range batch.ProposedNos {
		if no <= mgr.commitNo[id] {
			// committed previous partial order for node id, including partial number 0.
			mgr.logger.Debugf("[%d] haven't updated committed partial order for node %d", mgr.author, id)
			continue
		}

		pOrder := batch.Partials[id]

		if pOrder.Sequence() != no {
			return nil, fmt.Errorf("invalid partial order seqNo, proposedNo %d, partial seqNo %d", no, pOrder.Sequence())
		}

		if err := crypto.VerifyProofCerts(types.StringToBytes(pOrder.PreOrderDigest()), pOrder.QC, mgr.quorum); err != nil {
			return nil, fmt.Errorf("invalid high partial order received from %d: %s", batch.Author, err)
		}

		updated = true
	}

	if !updated {
		// haven't updated committed partial order.
		return nil, nil
	}

	var qStream types.QueryStream

	for id, no := range batch.ProposedNos {
		for {
			if no <= mgr.commitNo[id] {
				break
			}

			mgr.commitNo[id]++

			qIndex := types.QueryIndex{Author: id, SeqNo: mgr.commitNo[id]}
			qStream = append(qStream, qIndex)
		}
	}

	return qStream, nil
}
