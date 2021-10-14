package txmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sync"
)

// txManager is the implement of phalanx client, which is used receive transactions and generate phalanx commands.
type txManager struct {
	// mutex is used to deal with the concurrent problems of command-manager.
	mutex sync.Mutex

	//===================================== basic information ============================================

	// n indicates the member number of current cluster.
	n int

	// author indicates current node identifier.
	author uint64

	//====================================== command generator ============================================

	// seqNo is used to arrange the sequential number of generated commands.
	seqNo uint64

	// commandSize refer to the maximum count of transactions in one command.
	commandSize int

	// txSet is used to cache the commands current node has received.
	txSet []*protos.Transaction

	//======================================= external interfaces ==================================================

	// sender is used to send messages.
	sender external.TestSender

	// logger is used to print logs.
	logger external.Logger
}

func NewTxManager(n int, author uint64, commandSize int, sender external.TestSender, logger external.Logger) *txManager {
	return &txManager{n: n, author: author, commandSize: commandSize, sender: sender, logger: logger}
}

func (txMgr *txManager) ProcessTransaction(tx *protos.Transaction) {
	txMgr.mutex.Lock()
	defer txMgr.mutex.Unlock()

	txMgr.txSet = append(txMgr.txSet, tx)
	if len(txMgr.txSet) == txMgr.commandSize {
		txMgr.seqNo++
		command := types.GenerateCommand(txMgr.author, txMgr.seqNo, txMgr.txSet)
		txMgr.sender.BroadcastCommand(command)
		txMgr.logger.Infof("[%d] generate command %s", txMgr.author, command.Format())
		txMgr.txSet = nil
	}
}
