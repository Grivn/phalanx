package logmanager

import (
	"errors"
	"fmt"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/event"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/google/btree"
)

// subInstance is used to process the remote log, each instance would be used for one replica.
type subInstance struct {
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

	//======================================= internal modules =========================================

	// pTracker is used to record the partial orders from current sub instance node.
	pTracker *partialTracker

	//======================================= external tools ===========================================

	// sender is used to send votes to others.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func newSubInstance(author, id uint64, pTracker *partialTracker, sender external.NetworkService, logger external.Logger) *subInstance {
	logger.Infof("[%d] initiate the sub instance of order for replica %d", author, id)
	return &subInstance{
		author:   author,
		id:       id,
		trusted:  uint64(0),
		sequence: uint64(1),
		recorder: btree.New(2),
		pTracker: pTracker,
		sender:   sender,
		logger:   logger,
	}
}

func (si *subInstance) processPreOrder(pre *protos.PreOrder) error {
	si.logger.Infof("[%d] received a pre-order %s", si.author, pre.Format())

	if si.sequence > pre.Sequence {
		si.logger.Errorf("[%d] already voted on %d for replica %d", si.author, pre.Sequence, si.id)
		return nil
	}

	ev := &event.BtreeEvent{EventType: event.BTreeEventPreOrder, Sequence: pre.Sequence, Digest: pre.Digest, Event: pre}

	if si.recorder.Has(ev) {
		return si.processBTree()
	}

	if err := crypto.CheckDigest(pre); err != nil {
		return fmt.Errorf("invalid digest: %s", err)
	}

	si.recorder.ReplaceOrInsert(ev)

	return si.processBTree()
}

func (si *subInstance) processPartial(pOrder *protos.PartialOrder) error {
	si.logger.Infof("[%d] received a partial order %s", si.author, pOrder.Format())

	// verify the signatures of current received partial order.
	if err := crypto.VerifyProofCerts(types.StringToBytes(pOrder.PreOrderDigest()), pOrder.QC, si.quorum); err != nil {
		return fmt.Errorf("invalid order: %s", err)
	}

	ev := &event.BtreeEvent{EventType: event.BTreeEventOrder, Sequence: pOrder.PreOrder.Sequence, Digest: pOrder.PreOrder.Digest, Event: pOrder}
	si.recorder.ReplaceOrInsert(ev)

	return si.processBTree()
}

func (si *subInstance) processBTree() error {
	item := si.recorder.Min()
	if item == nil {
		return nil
	}
	ev := item.(*event.BtreeEvent)

	switch ev.EventType {
	case event.BTreeEventPreOrder:
		if ev.Sequence != si.sequence {
			si.logger.Debugf("[%d] sub-instance for node %d needs sequence %d", si.author, si.id, si.sequence)
			return nil
		}

		// parsing the event info
		pre := ev.Event.(*protos.PreOrder)

		// generate the signature for current pre-order
		sig, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(si.author))
		if err != nil {
			return fmt.Errorf("signer failed: %s", err)
		}

		// generate and send vote to the pre-order author
		vote := &protos.Vote{Author: si.author, Digest: pre.Digest, Certification: sig}
		si.logger.Infof("[%d] voted %s for %s", si.author, vote.Format(), pre.Format())

		cm, err := protos.PackVote(vote, pre.Author)
		if err != nil {
			return fmt.Errorf("generate consensus message error: %s", err)
		}
		si.sender.UnicastPCM(cm)
		return nil

	case event.BTreeEventOrder:
		si.logger.Infof("[%d] process partial order event", si.author)

		pOrder := ev.Event.(*protos.PartialOrder)

		// verify the validation between current partial order and highest partial order.
		if err := si.checkHighestOrder(pOrder); err != nil {
			return fmt.Errorf("check higest order failed, %s", err)
		}

		// record partial order with partial tracker.
		si.pTracker.recordPartial(pOrder)

		// update the highest partial order for current sub instance.
		si.updateHighestOrder(pOrder)

		si.recorder.Delete(ev)
		si.sequence++

		if err := si.processBTree(); err != nil {
			return err
		}
		return nil

	default:
		return errors.New("invalid event type")
	}
}

func (si *subInstance) updateHighestOrder(pOrder *protos.PartialOrder) {
	si.highPartialOrder = pOrder
}

func (si *subInstance) checkHighestOrder(pOrder *protos.PartialOrder) error {
	if pOrder.Sequence() == 1 {
		// the first partial order for current sub instance.
		return nil
	}

	if si.highPartialOrder == nil {
		// we don't have a partial order here, reject it.
		return fmt.Errorf("nil highest order")
	}

	if si.highPartialOrder.PreOrderDigest() != pOrder.ParentOrder().Digest {
		return fmt.Errorf("invalid parent order digest, expect %s, received %s", si.highPartialOrder.PreOrderDigest(), pOrder.ParentOrder().Digest)
	}

	return nil
}
