package subinstance

import (
	"errors"
	"fmt"
	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/internal"

	"github.com/Grivn/phalanx/common/event"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/google/btree"
)

type subInstance struct {
	// author is the local node's identifier
	author uint64

	// id indicates which node the current request-pool is maintained for
	id uint64

	// quorum indicates the set size for invalid signatures
	quorum int

	// sequence is the preferred number for the next request
	sequence uint64

	// recorder is used to track the pre-order/order messages
	recorder *btree.BTree

	// sender is used to send votes to others
	sender external.Network

	sequential internal.SequencePool

	// logger is used to print logs
	logger external.Logger
}

func newSubInstance(author, id uint64, preC chan *protos.PreOrder, sender external.Network, logger external.Logger) *subInstance {
	logger.Noticef("replica %d init the sub instance of order for replica %d", author, id)
	return &subInstance{
		author:   author,
		id:       id,
		sequence: uint64(0),

		recorder: btree.New(2),

		sender: sender,
		logger: logger,
	}
}

func (si *subInstance) SubID() uint64 {
	return si.id
}

func (si *subInstance) ProcessPreOrder(pre *protos.PreOrder) error {
	si.logger.Infof("replica %d received a pre-order message from replica %d, hash %s", si.author, pre.Author, pre.Digest)

	ev := &event.BtreeEvent{EventType: event.BTreeEventPreOrder, Seq: pre.Sequence, Digest: pre.Digest, Event: pre}

	if si.recorder.Has(ev) {
		return si.processBTree()
	}

	if err := pre.CheckDigest(); err != nil {
		return fmt.Errorf("invalid digest: %s", err)
	}

	si.recorder.ReplaceOrInsert(ev)

	return si.processBTree()
}

func (si *subInstance) ProcessOrder(order *protos.Order) error {
	si.logger.Infof("replica %d received a order message", si.author)

	if err := order.Verify(si.quorum); err != nil {
		return fmt.Errorf("invalid order: %s", err)
	}

	ev := &event.BtreeEvent{EventType: event.BTreeEventOrder, Seq: order.PreOrder.Sequence, Digest: order.PreOrder.Digest, Event: order}
	si.recorder.ReplaceOrInsert(ev)

	return si.processBTree()
}

func (si *subInstance) processBTree() error {
	ev := si.recorder.Min().(*event.BtreeEvent)

	switch ev.EventType {
	case event.BTreeEventPreOrder:
		// parsing the event info
		pre := ev.Event.(*protos.PreOrder)

		// generate the signature for current pre-order
		sig, err := crypto.PrivSign(types.StringToBytes(pre.Digest), int(si.author))
		if err != nil {
			return fmt.Errorf("signer failed: %s", err)
		}

		// generate and send vote to the pre-order author
		vote := &protos.Vote{Author: si.author, Digest: pre.Digest, Certification: sig}
		si.sender.SendVote(vote, pre.Author)

	case event.BTreeEventOrder:
		order := ev.Event.(*protos.Order)

		si.sequential.InsertQuorumCert(&protos.QuorumCert{Order: order})
	default:
		return errors.New("invalid event type")
	}

	return nil
}
