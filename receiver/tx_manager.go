package receiver

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

// txManager is the implement of phalanx client, which is used receive transactions and generate phalanx commands.
type txManager struct {
	//===================================== basic information ============================================

	// author indicates current node identifier.
	author uint64

	//====================================== command generator ============================================

	// proposers are the proposer instances to generate commands.
	proposers map[uint64]*proposerImpl

	// txC is used to submit transactions.
	txC chan<- *protos.Transaction

	//======================================= external interfaces ==================================================

	// logger is used to print logs.
	logger external.Logger
}

func NewTxManager(conf Config) api.Proposer {
	proposers := make(map[uint64]*proposerImpl)

	txC := make(chan *protos.Transaction)

	base := int(conf.Author-1) * conf.Multi

	for i := base; i < base+conf.Multi; i++ {
		id := uint64(i + 1)
		proposer := newProposer(id, txC, conf)
		proposers[id] = proposer
	}

	return &txManager{author: conf.Author, proposers: proposers, txC: txC, logger: conf.Logger}
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
