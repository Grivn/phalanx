package txmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"time"
)

// txManager is the implement of phalanx client, which is used receive transactions and generate phalanx commands.
type txManager struct {
	//===================================== basic information ============================================

	// author indicates current node identifier.
	author uint64

	//====================================== command generator ============================================

	//
	txCount int32

	//
	memSize int32

	// proposers are the proposer instances to generate commands.
	proposers map[uint64]*proposerImpl

	// txC is used to submit transactions.
	txC chan<- *protos.Transaction

	//======================================= external interfaces ==================================================

	// logger is used to print logs.
	logger external.Logger
}

func NewTxManager(interval int, duration time.Duration, multi int, author uint64, commandSize int, memSize int, sender external.NetworkService, logger external.Logger) internal.TxManager {
	proposers := make(map[uint64]*proposerImpl)

	txC := make(chan *protos.Transaction)

	base := int(author-1)*multi

	fronts := newInnerFronts()

	for i:=base; i<base+multi; i++ {
		id := uint64(i+1)

		readFront := id-1

		if i == base {
			readFront = uint64(base+multi+1)
		}

		proposer := newProposer(id, commandSize, memSize, txC, sender, logger, fronts, interval, readFront, duration)

		proposers[id] = proposer
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

func (txMgr *txManager) Reply(command *protos.Command) {
	proposer, ok := txMgr.proposers[command.Author]
	if !ok {
		return
	}
	proposer.reply(command)
}

func (txMgr *txManager) ProcessTransaction(tx *protos.Transaction) {
	txMgr.txC <- tx
}
