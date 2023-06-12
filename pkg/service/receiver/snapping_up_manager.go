package receiver

import (
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/external"
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

func NewSnappingUpManager(conf config.PhalanxConf, sender external.Sender, logger external.Logger) api.Proposer {
	buyers := make(map[uint64]*buyerImpl)
	base := int(conf.NodeID-1) * conf.Multi
	for i := base; i < base+conf.Multi; i++ {
		id := uint64(i + 1)
		buyer := newBuyer(id, conf, sender, logger)
		buyers[id] = buyer
	}

	return &snappingUpManagerImpl{
		author: conf.NodeID,
		buyers: buyers,
		logger: logger,
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
