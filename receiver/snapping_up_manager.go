package receiver

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type snappingUpManagerImpl struct {
	// author indicates current node identifier.
	author uint64

	// buyers are the proposer instances to generate commands.
	buyers map[uint64]*buyerImpl

	// timer is used to
	timer *localTimer

	// committedItems
	committedItems map[uint64]bool

	// shoppingCarts is used to collect the number of items each buyer has bought.
	shoppingCarts map[uint64][]uint64

	logger external.Logger
}

func NewSnappingUpManager(conf Config) api.Proposer {
	buyers := make(map[uint64]*buyerImpl)
	base := int(conf.Author-1) * conf.Multi
	for i := base; i < base+conf.Multi; i++ {
		id := uint64(i + 1)
		buyer := newBuyer(id, conf)
		buyers[id] = buyer
	}

	return &snappingUpManagerImpl{
		author: conf.Author,
		buyers: buyers,
		logger: conf.Logger,
	}
}

func (su *snappingUpManagerImpl) Run() {
	for _, buyer := range su.buyers {
		go buyer.run()
	}
}

func (su *snappingUpManagerImpl) Quit() {
	for _, buyer := range su.buyers {
		buyer.quit()
	}
}

func (su *snappingUpManagerImpl) ProcessTransaction(tx *protos.Transaction) {
	// no use.
}

func (su *snappingUpManagerImpl) CommitResult(itemNo uint64, buyer uint64) {

}
