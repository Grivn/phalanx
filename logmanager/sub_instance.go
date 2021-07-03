package logmanager

import (
	"errors"
	"fmt"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/event"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"

	"github.com/google/btree"
)

// subInstance is used to process the remote log, each instance would be used for one replica.
type subInstance struct {
	// author is the local node's identifier.
	author uint64

	// id indicates which node the current request-pool is maintained for.
	id uint64

	// quorum indicates the set size for invalid signatures.
	quorum int

	// sequence is the preferred seqNo for the next request.
	sequence uint64

	// recorder is used to track the pre-order/order messages.
	recorder *btree.BTree

	// sender is used to send votes to others.
	sender external.NetworkService

	// sp is the seq-pool module for phalanx.
	sp internal.SequencePool

	// logger is used to print logs.
	logger external.Logger
}

func newSubInstance(author, id uint64, sp internal.SequencePool, sender external.NetworkService, logger external.Logger) *subInstance {
	logger.Infof("[%d] initiate the sub instance of order for replica %d", author, id)
	return &subInstance{
		author:   author,
		id:       id,
		sequence: uint64(1),
		recorder: btree.New(2),
		sp:       sp,
		sender:   sender,
		logger:   logger,
	}
}

func (si *subInstance) processPreOrder(pre *protos.PreOrder) error {
	si.logger.Infof("[%d] received a pre-order %s", si.author, pre.Format())

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

		si.sequence++
		return nil

	case event.BTreeEventOrder:
		si.logger.Infof("[%d] process partial order event", si.author)

		pOrder := ev.Event.(*protos.PartialOrder)

		if err := si.sp.InsertPartialOrder(pOrder); err != nil {
			si.logger.Errorf("[%d] insert failed: %s", si.author, err)
		}

		si.recorder.Delete(ev)

		if err := si.processBTree(); err != nil {
			return err
		}
		return nil

	default:
		return errors.New("invalid event type")
	}
}
