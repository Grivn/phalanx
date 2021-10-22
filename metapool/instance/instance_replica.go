package instance

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/event"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"

	"github.com/google/btree"
)

// replicaInstance is used to process the remote log, each instance would be used for one replica.
type replicaInstance struct {
	//===================================== basic information =========================================

	// author is the local node's identifier.
	author uint64

	// id indicates which node the current request-pool is maintained for.
	id uint64

	//==================================== sub-chain management =============================================

	// quorum indicates the set size for invalid signatures.
	quorum int

	// trusted indicates the highest verified seqNo.
	trusted uint64

	// mutex is used to process concurrency problems for highPartialOrder.
	mutex sync.RWMutex

	// highPartialOrder indicates the highest partial order which has been verified by cluster validators.
	highPartialOrder *protos.PartialOrder

	// sequence is the preferred seqNo for the next request.
	sequence uint64

	// recorder is used to track the pre-order/order messages.
	recorder *btree.BTree

	//====================================== communication channel ===========================================

	preC chan *protos.PreOrder

	partialC chan *protos.PartialOrder

	closeC chan bool

	//======================================= internal modules =========================================

	// pTracker is used to record the partial orders from current sub instance node.
	pTracker internal.PartialTracker

	//======================================= external tools ===========================================

	// sender is used to send votes to others.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func NewReplicaInstance(author, id uint64, pTracker internal.PartialTracker, sender external.NetworkService, logger external.Logger) internal.ReplicaInstance {
	logger.Infof("[%d] initiate the sub instance of order for replica %d", author, id)
	return &replicaInstance{
		author:   author,
		id:       id,
		trusted:  uint64(0),
		sequence: uint64(1),
		recorder: btree.New(2),
		pTracker: pTracker,
		preC:     make(chan *protos.PreOrder),
		partialC: make(chan *protos.PartialOrder),
		closeC:   make(chan bool),
		sender:   sender,
		logger:   logger,
	}
}

func (ri *replicaInstance) Run() {
	go ri.listener()
}

func (ri *replicaInstance) Close() {
	select {
	case <-ri.closeC:
	default:
		close(ri.closeC)
	}
}

func (ri *replicaInstance) ReceivePreOrder(pre *protos.PreOrder) {
	ri.preC <- pre
}

func (ri *replicaInstance) ReceivePartial(pOrder *protos.PartialOrder) {
	ri.partialC <- pOrder
}

func (ri *replicaInstance) GetHighOrder() *protos.PartialOrder {
	return ri.getHighOrder()
}

//===================================== coroutine worker =====================================================

func (ri *replicaInstance) listener() {
	for {
		select {
		case <-ri.closeC:
			return
		case pre := <-ri.preC:
			if err := ri.processPreOrder(pre); err != nil {
				ri.logger.Errorf("[%d] process pre order error: %s", ri.author, err)
			}
		case pOrder := <-ri.partialC:
			if err := ri.processPartial(pOrder); err != nil {
				ri.logger.Errorf("[%d] process partial order error: %s", ri.author, err)
			}
		}
	}
}

func (ri *replicaInstance) processPreOrder(pre *protos.PreOrder) error {
	ri.logger.Infof("[%d] received a pre-order %s", ri.author, pre.Format())

	if ri.sequence > pre.Sequence {
		ri.logger.Errorf("[%d] already voted on %d for replica %d", ri.author, pre.Sequence, ri.id)
		return nil
	}

	ev := &event.OrderEvent{Status: event.OrderStatusPreOrder, Sequence: pre.Sequence, Digest: pre.Digest, Event: pre}

	if ri.recorder.Has(ev) {
		return ri.processBTree()
	}

	if err := crypto.CheckDigest(pre); err != nil {
		return fmt.Errorf("invalid digest: %s", err)
	}

	ri.recorder.ReplaceOrInsert(ev)

	return ri.processBTree()
}

func (ri *replicaInstance) processPartial(pOrder *protos.PartialOrder) error {
	ri.logger.Infof("[%d] received a partial order %s", ri.author, pOrder.Format())

	// verify the signatures of current received partial order.
	if err := crypto.VerifyProofCerts(types.StringToBytes(pOrder.PreOrderDigest()), pOrder.QC, ri.quorum); err != nil {
		return fmt.Errorf("invalid order: %s", err)
	}

	ev := &event.OrderEvent{Status: event.OrderStatusQuorumVerified, Sequence: pOrder.PreOrder.Sequence, Digest: pOrder.PreOrder.Digest, Event: pOrder}
	ri.recorder.ReplaceOrInsert(ev)

	return ri.processBTree()
}

func (ri *replicaInstance) processBTree() error {
	item := ri.recorder.Min()
	if item == nil {
		return nil
	}
	ev := item.(*event.OrderEvent)

	switch ev.Status {
	case event.OrderStatusPreOrder:
		if ev.Sequence != ri.sequence {
			ri.logger.Debugf("[%d] sub-instance for node %d needs sequence %d", ri.author, ri.id, ri.sequence)
			return nil
		}

		// parsing the event info
		pre := ev.Event.(*protos.PreOrder)

		// generate the signature for current pre-order
		sig, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(ri.author))
		if err != nil {
			return fmt.Errorf("signer failed: %s", err)
		}

		// generate and send vote to the pre-order author
		vote := &protos.Vote{Author: ri.author, Digest: pre.Digest, Certification: sig}
		ri.logger.Infof("[%d] voted %s for %s", ri.author, vote.Format(), pre.Format())

		cm, err := protos.PackVote(vote, pre.Author)
		if err != nil {
			return fmt.Errorf("generate consensus message error: %s", err)
		}
		ri.sender.UnicastPCM(cm)
		return nil

	case event.OrderStatusQuorumVerified:
		ri.logger.Infof("[%d] process partial order event", ri.author)

		pOrder := ev.Event.(*protos.PartialOrder)

		// verify the validation between current partial order and highest partial order.
		if err := ri.checkHighOrder(pOrder); err != nil {
			return nil
			//return fmt.Errorf("check higest order failed, %s", err)
		}

		// record partial order with partial tracker.
		ri.pTracker.RecordPartial(pOrder)

		// update the highest partial order for current sub instance.
		ri.updateHighOrder(pOrder)

		ri.recorder.Delete(ev)
		ri.sequence++

		if err := ri.processBTree(); err != nil {
			return err
		}
		return nil

	default:
		return errors.New("invalid event type")
	}
}

//========================================== high order management ==============================================

func (ri *replicaInstance) getHighOrder() *protos.PartialOrder {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()

	return ri.highPartialOrder
}

func (ri *replicaInstance) updateHighOrder(pOrder *protos.PartialOrder) {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	ri.highPartialOrder = pOrder
}

func (ri *replicaInstance) checkHighOrder(pOrder *protos.PartialOrder) error {
	if pOrder.Sequence() == 1 {
		// the first partial order for current sub instance.
		return nil
	}

	if ri.highPartialOrder == nil {
		// we don't have a partial order here, reject it.
		return fmt.Errorf("nil highest order")
	}

	if ri.highPartialOrder.PreOrderDigest() != pOrder.ParentDigest() {
		return fmt.Errorf("invalid parent order digest, expect %s, received %s, pOrder %s", ri.highPartialOrder.PreOrderDigest(), pOrder.ParentDigest(), pOrder.Format())
	}

	return nil
}
