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

	//====================================== command generators ============================================

	// transactionC is used to receive transactions.
	transactionC chan <- *protos.Transaction

	// proposers are the exact generators of command.
	proposers []*subProposer

	//======================================= external interfaces ==================================================

	// logger is used to print logs.
	logger external.Logger
}

func NewTxManager(n int, author uint64, commandSize int, concurrency int, sender external.NetworkService, logger external.Logger) internal.TxManager {
	var proposers []*subProposer
	transactionC := make(chan *protos.Transaction)

	if concurrency <= 0 {
		logger.Errorf("[%d] invalid concurrency count %d, make concurrency legal", author, concurrency)
		concurrency = 1
	}

	for i:=0; i<concurrency; i++ {
		id := author + uint64(n*i)
		proposer := newSubProposer(author, id, commandSize, transactionC, sender, logger)
		proposers = append(proposers, proposer)
	}

	return &txManager{author: author, transactionC: transactionC, proposers: proposers, logger: logger}
}

func (txMgr *txManager) Run() {
	for _, proposer := range txMgr.proposers {
		proposer.run()
	}
}

func (txMgr *txManager) Close() {
	for _, proposer := range txMgr.proposers {
		proposer.close()
	}
}

func (txMgr *txManager) ProcessTransaction(tx *protos.Transaction) {
	txMgr.transactionC <- tx
}
