package instance

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/event"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/google/btree"
)

// replicaInstance is used to process the remote log, each instance would be used for one replica.
type replicaInstance struct {
	// mutex is used to resolve concurrency problems.
	mutex sync.RWMutex

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

	// highPartialOrder indicates the highest partial order which has been verified by cluster validators.
	highPartialOrder *protos.PartialOrder

	// sequence is the preferred seqNo for the next request.
	sequence uint64

	// recorder is used to track the pre-order/order messages.
	recorder *btree.BTree

	// voted is used to record the latest no. we have verified.
	voted uint64

	//======================================= internal modules =========================================

	// pTracker is used to record the partial orders from current sub instance node.
	pTracker api.PartialTracker

	//==================================== crypto management =============================================

	// crypto is used to generate/verify certificates.
	crypto api.Crypto

	//======================================= external tools ===========================================

	// sender is used to send votes to others.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func NewReplicaInstance(author, id uint64, pTracker api.PartialTracker, crypto api.Crypto,
	sender external.NetworkService, logger external.Logger) api.ReplicaInstance {
	logger.Infof("[%d] initiate the sub instance of order for replica %d", author, id)
	return &replicaInstance{
		author:   author,
		id:       id,
		trusted:  uint64(0),
		sequence: uint64(1),
		voted:    uint64(0),
		recorder: btree.New(2),
		pTracker: pTracker,
		crypto:   crypto,
		sender:   sender,
		logger:   logger,
	}
}

func (ri *replicaInstance) GetHighOrder() *protos.PartialOrder {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()
	return ri.highPartialOrder
}

func (ri *replicaInstance) ReceivePreOrder(pre *protos.PreOrder) error {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	ri.logger.Infof("[%d] received a pre-order %s", ri.author, pre.Format())

	if ri.sequence > pre.Sequence {
		ri.logger.Errorf("[%d] already voted on %d for replica %d", ri.author, pre.Sequence, ri.id)
		return nil
	}

	ev := &event.OrderEvent{Status: event.OrderStatusPreOrder, Sequence: pre.Sequence, Digest: pre.Digest, Event: pre}

	if ri.recorder.Has(ev) {
		return ri.processBTree()
	}

	if err := types.CheckPreOrderDigest(pre); err != nil {
		return fmt.Errorf("invalid digest: %s", err)
	}

	ri.recorder.ReplaceOrInsert(ev)

	return ri.processBTree()
}

func (ri *replicaInstance) ReceivePartial(pOrder *protos.PartialOrder) error {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	ri.logger.Infof("[%d] received a partial order %s", ri.author, pOrder.Format())

	// verify the signatures of current received partial order.
	if err := ri.crypto.VerifyProofCerts(types.StringToBytes(pOrder.PreOrderDigest()), pOrder.QC, ri.quorum); err != nil {
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
		sig, err := ri.crypto.PrivateSign(types.StringToBytes(pre.Digest))
		if err != nil {
			return fmt.Errorf("signer failed: %s", err)
		}

		// generate and send vote to the pre-order author
		vote := &protos.Vote{Author: ri.author, Digest: pre.Digest, Certification: sig}
		ri.logger.Infof("[%d] voted %s for %s", ri.author, vote.Format(), pre.Format())
		ri.recorder.Delete(item)

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
		if err := ri.checkHighestOrder(pOrder); err != nil {
			return nil
			//return fmt.Errorf("check higest order failed, %s", err)
		}

		// record partial order with partial tracker.
		ri.pTracker.RecordPartial(pOrder)

		// update the highest partial order for current sub instance.
		ri.updateHighestOrder(pOrder)

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

func (ri *replicaInstance) updateHighestOrder(pOrder *protos.PartialOrder) {
	ri.highPartialOrder = pOrder
}

func (ri *replicaInstance) checkHighestOrder(pOrder *protos.PartialOrder) error {
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
