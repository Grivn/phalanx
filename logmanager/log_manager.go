package logmanager

import (
	"fmt"
	"github.com/google/btree"
	"sync"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"

	"github.com/gogo/protobuf/proto"
)

type logManager struct {
	// mutex is used to deal with the concurrent problems of log-manager.
	mutex sync.Mutex

	// author is the identifier for current node.
	author uint64

	// quorum is the legal size for current node.
	quorum int

	// sequence is the the target for local-log.
	sequence uint64

	// aggMap is used to generate aggregated-certificates.
	aggMap map[string]*protos.PartialOrder

	// subs is the module for us to process consensus messages for participates.
	subs map[uint64]*subInstance

	// sender is used to send consensus message into network.
	sender external.NetworkService

	// clients is used to track the commands send from them.
	clients map[uint64]*clientInfo

	//
	commandC chan *protos.Command

	//
	closeC chan bool

	// logger is used to print logs.
	logger external.Logger
}

type clientInfo struct {
	// mutex
	mutex sync.Mutex

	// author
	author uint64

	// id
	id uint64

	// proposedNo
	proposedNo uint64

	// commands
	commands *btree.BTree

	// commandC
	commandC chan *protos.Command

	// logger
	logger external.Logger
}

func newClient(author, id uint64, commandC chan *protos.Command, logger external.Logger) *clientInfo {
	logger.Infof("[%d] initiate manager for client %d", author, id)
	return &clientInfo{author: author, id: id, proposedNo: uint64(0), commands: btree.New(2), commandC: commandC, logger: logger}
}

func (client *clientInfo) append(command *protos.Command) {
	client.logger.Debugf("[%d] received command %s", client.author, command.Format())
	client.commands.ReplaceOrInsert(command)
}

func (client *clientInfo) minCommand() *protos.Command {
	item := client.commands.Min()
	if item == nil {
		return nil
	}

	command, ok := item.(*protos.Command)
	if !ok {
		return nil
	}

	if command.Sequence == client.proposedNo+1 {
		client.commands.Delete(item)
		client.proposedNo++
		return command
	}
	return nil
}

func NewLogManager(n int, author uint64, sp internal.SequencePool, sender external.NetworkService, logger external.Logger) *logManager {
	logger.Infof("[%d] initiate log manager, replica count %d", author, n)
	subs := make(map[uint64]*subInstance)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		subs[id] = newSubInstance(author, id, sp, sender, logger)
	}

	return &logManager{
		quorum:   types.CalculateQuorum(n),
		author:   author,
		sequence: uint64(0),
		aggMap:   make(map[string]*protos.PartialOrder),
		subs:     subs,
		clients:  make(map[uint64]*clientInfo),
		commandC: make(chan *protos.Command),
		closeC:   make(chan bool),
		sender:   sender,
		logger:   logger,
	}
}

//===============================================================
//                 Processor for Local Logs
//===============================================================

func (mgr *logManager) ProcessCommand(command *protos.Command) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	client, ok := mgr.clients[command.Author]
	if !ok {
		client = newClient(mgr.author, command.Author, mgr.commandC, mgr.logger)
		mgr.clients[command.Author] = client
	}
	client.append(command)

	return mgr.tryGeneratePreOrder(command.Author)
}

// generatePreOrder is used to process command received from clients.
// We would like to assign a sequence number for such a command and generate a pre-order message.
func (mgr *logManager) tryGeneratePreOrder(id uint64) error {
	command := mgr.clients[id].minCommand()

	if command == nil {
		return nil
	}

	mgr.sequence++

	pre := protos.NewPreOrder(mgr.author, mgr.sequence, command)
	payload, err := proto.Marshal(pre)
	if err != nil {
		mgr.logger.Errorf("Marshal Error: %v", err)
		return err
	}
	pre.Digest = types.CalculatePayloadHash(payload, 0)

	// generate self-signature for current pre-order
	signature, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(mgr.author))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}

	// init the order message in aggregate map and assign self signature
	mgr.aggMap[pre.Digest] = protos.NewPartialOrder(pre)
	mgr.aggMap[pre.Digest].QC.Certs[mgr.author] = signature

	mgr.logger.Infof("[%d] generate pre-order %s", mgr.author, pre.Format())

	cm, err := protos.PackPreOrder(pre)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	mgr.sender.BroadcastPCM(cm)
	return mgr.tryGeneratePreOrder(id)
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
