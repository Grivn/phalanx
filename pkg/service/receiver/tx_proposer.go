package receiver

import (
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
)

type proposerImpl struct {
	//===================================== basic information ============================================

	// author indicates current node identifier.
	author uint64

	//====================================== command generator ============================================

	// seqNo is used to arrange the sequential number of generated commands.
	seqNo uint64

	// commandSize refer to the maximum count of transactions in one command.
	commandSize int

	// txSet is used to cache the commands current node has received.
	txSet []*protos.Transaction

	// txC is used to receive transactions.
	txC <-chan *protos.Transaction

	// closeC is used to close proposer listener.
	closeC chan bool

	//
	memSize int32

	//======================================= external interfaces ==================================================

	// sender is used to send messages.
	sender external.Sender

	// logger is used to print logs.
	logger external.Logger

	//
	selected uint64
}

func newProposer(author uint64, txC chan *protos.Transaction, conf config.PhalanxConf,
	sender external.Sender, logger external.Logger) *proposerImpl {
	return &proposerImpl{
		author:      author,
		commandSize: conf.CommandSize,
		txC:         txC,
		closeC:      make(chan bool),
		sender:      sender,
		logger:      logger,
		memSize:     int32(conf.MemSize),
		selected:    conf.Selected,
	}
}

func (p *proposerImpl) run() {
	for {
		select {
		case <-p.closeC:
			return
		case tx := <-p.txC:
			p.processTx(tx)
		}
	}
}

func (p *proposerImpl) quit() {
	select {
	case <-p.closeC:
	default:
		close(p.closeC)
	}
}

func (p *proposerImpl) processTx(tx *protos.Transaction) {
	if p.selected != uint64(0) && p.author < p.selected {
		return
	}

	p.txSet = append(p.txSet, tx)
	if len(p.txSet) == p.commandSize {
		p.seqNo++
		command := types.GenerateCommand(p.author, p.seqNo, p.txSet)
		p.sender.BroadcastCommand(command)
		p.logger.Infof("[%d] generate command %s", p.author, command.Format())
		p.txSet = nil
	}
}
