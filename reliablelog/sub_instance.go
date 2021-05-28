package reliablelog

import (
	"sync"

	commonProto "github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/google/btree"
)

type subInstance struct {
	// author is the local node's identifier
	author uint64

	// id indicates which node the current request-pool is maintained for
	id uint64

	// mutex is used to resolve concurrent problems with sequence number
	mutex sync.Mutex

	// sequence is the preferred number for the next request
	sequence uint64

	// recorder is used to track the pre-order/order messages
	recorder *btree.BTree

	// sender is used to send votes to others
	sender external.Network

	// logger is used to print logs
	logger external.Logger
}

type btreeOrder struct {
	seq     uint64
	pre     *commonProto.PreOrder
	ordered *commonProto.Order
}

func (bo *btreeOrder) Less(item btree.Item) bool {
	return bo.seq < (item.(*btreeOrder)).seq
}

func newSubInstance(author, id uint64, preC chan *commonProto.PreOrder, sender external.Network, logger external.Logger) *subInstance {
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

// todo start a listener for pre-order

// todo start a listener for order

// todo process concurrent problem for sequence number

func (si *subInstance) processPreOrder(pre *commonProto.PreOrder) {
	si.logger.Infof("replica %d received a pre-order message from replica %d, hash %s", si.author, pre.Author, pre.Digest)

	si.preRecorder.ReplaceOrInsert(&btreePreOrder{seq: pre.Sequence, value: pre})
}

func (si *subInstance) processOrder(order *commonProto.Order) {
	si.logger.Infof("replica %d received a order message", si.author)


}

func (si *subInstance) currentSequence() uint64 {
	si.mutex.Lock()
	defer si.mutex.Unlock()
}
