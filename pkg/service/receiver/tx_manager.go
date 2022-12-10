package receiver

import (
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/external"
)

// txManager is the implement of phalanx client, which is used receive transactions and generate phalanx commands.
type txManager struct {
	//===================================== basic information ============================================

	// nodeID indicates current node identifier.
	nodeID uint64

	//====================================== command generator ============================================

	// proposers are the proposer instances to generate commands.
	proposers map[uint64]*proposerImpl

	// txC is used to submit transactions.
	txC chan<- *protos.Transaction

	//======================================= external interfaces ==================================================

	// logger is used to print logs.
	logger external.Logger
}

func NewTxManager(conf config.PhalanxConf, sender external.Sender, logger external.Logger) api.Proposer {
	proposers := make(map[uint64]*proposerImpl)

	txC := make(chan *protos.Transaction)

	base := int(conf.NodeID-1) * conf.Multi

	for i := base; i < base+conf.Multi; i++ {
		id := uint64(i + 1)
		proposer := newProposer(id, txC, conf, sender, logger)
		proposers[id] = proposer
	}

	return &txManager{nodeID: conf.NodeID, proposers: proposers, txC: txC, logger: logger}
}

func (txMgr *txManager) Run() {
	for _, proposer := range txMgr.proposers {
		go proposer.run()
	}
}

func (txMgr *txManager) Quit() {
	for _, proposer := range txMgr.proposers {
		proposer.quit()
	}
}

func (txMgr *txManager) ProcessTransaction(tx *protos.Transaction) {
	txMgr.txC <- tx
}
