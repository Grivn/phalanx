package txmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

// txManager is the implement of phalanx client, which is used receive transactions and generate phalanx commands.
type txManager struct {
	//===================================== basic information ============================================

	// author indicates current node identifier.
	author uint64

	//====================================== command generator ============================================

	// proposers are the proposer instances to generate commands.
	proposers []*proposerImpl

	// txC is used to submit transactions.
	txC chan<- *protos.Transaction

	//======================================= external interfaces ==================================================

	// logger is used to print logs.
	logger external.Logger
}

func NewTxManager(multi int, author uint64, commandSize int, sender external.NetworkService, logger external.Logger) internal.TxManager {
	var proposers []*proposerImpl

	txC := make(chan *protos.Transaction)

	base := int(author-1)*multi

	for i:=base; i<base+multi; i++ {
		id := uint64(i+1)

		proposer := newProposer(id, commandSize, txC, sender, logger)

		proposers = append(proposers, proposer)
	}

	return &txManager{author: author, proposers: proposers, txC: txC, logger: logger}
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
