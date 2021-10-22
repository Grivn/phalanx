package txmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type subProposer struct {
	//===================================== basic information ============================================

	// author indicates current replica identifier.
	author uint64

	// id indicates current proposer identifier.
	id uint64

	//====================================== command generator ============================================

	// seqNo is used to arrange the sequential number of generated commands.
	seqNo uint64

	// commandSize refer to the maximum count of transactions in one command.
	commandSize int

	// txSet is used to cache the commands current node has received.
	txSet []*protos.Transaction

	//======================================= communication channel ================================================

	// transactionC is used to receive phalanx transactions.
	transactionC chan *protos.Transaction

	// closeC is used to close the coroutine of current proposer.
	closeC chan bool

	//======================================= external interfaces ==================================================

	// sender is used to send messages.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func newSubProposer(author uint64, id uint64, commandSize int, transactionC chan *protos.Transaction, sender external.NetworkService, logger external.Logger) *subProposer {
	logger.Infof("[%d] initiate proposer %d", author, id)
	return &subProposer{
		author:       author,
		id:           id,
		commandSize:  commandSize,
		transactionC: transactionC,
		closeC:       make(chan bool),
		sender:       sender,
		logger:       logger,
	}
}

func (sub *subProposer) run() {
	go sub.listener()
}

func (sub *subProposer) close() {
	select {
	case <-sub.closeC:
	default:
		close(sub.closeC)
	}
}

func (sub *subProposer) listener() {
	for {
		select {
		case <-sub.closeC:
			return
		case tx := <-sub.transactionC:
			sub.txSet = append(sub.txSet, tx)
			if len(sub.txSet) == sub.commandSize {
				// advance the assigned command seqNo.
				sub.seqNo++

				// generate command and broadcast.
				command := types.GenerateCommand(sub.id, sub.seqNo, sub.txSet)
				sub.sender.BroadcastCommand(command)
				sub.logger.Infof("[%d] generate command %s", sub.author, command.Format())

				// set nil transaction set.
				sub.txSet = nil
			}
		}
	}
}
