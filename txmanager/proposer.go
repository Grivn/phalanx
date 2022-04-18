package txmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sync/atomic"
	"time"
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
	intervalC <-chan *protos.Command

	//
	interval int

	//
	duration time.Duration

	//
	txCount int32

	//
	memSize int32

	//======================================= external interfaces ==================================================

	// sender is used to send messages.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger

	//
	front *innerFronts

	//
	readFront uint64

	//
	timer *localTimer

	//
	selected uint64
}

func newProposer(author uint64, commandSize int, memSize int, txC chan *protos.Transaction,
	sender external.NetworkService, logger external.Logger,
	front *innerFronts, interval int, readFront uint64, duration time.Duration, selected uint64) *proposerImpl {
	intervalC := make(chan *protos.Command)
	return &proposerImpl{
		author:      author,
		commandSize: commandSize,
		txC:         txC,
		closeC:      make(chan bool),
		intervalC:   intervalC,
		sender:      sender,
		logger:      logger,
		memSize:     int32(memSize),
		front:       front,
		interval:    interval,
		readFront:   readFront,
		duration:    duration,
		timer:       newLocalTimer(author, intervalC, duration, logger),
		selected:    selected,
	}
}

func (p *proposerImpl) run() {
	for {
		select {
		case <-p.closeC:
			return
		case tx := <-p.txC:
			p.processTx(tx)
		case command := <-p.intervalC:
			p.front.update(command)
		}
	}
}

func (p *proposerImpl) quit() {
	p.timer.stopTimer()
	select {
	case <-p.closeC:
	default:
		close(p.closeC)
	}
}

func (p *proposerImpl) reply(command *protos.Command) {
	count := len(command.Content)
	atomic.AddInt32(&p.txCount, 0-int32(count))
}

func (p *proposerImpl) processTx(tx *protos.Transaction) {
	if p.selected != uint64(0) && p.author > p.selected {
		return
	}
	if atomic.LoadInt32(&p.txCount) == p.memSize {
		return
	}

	p.txSet = append(p.txSet, tx)
	atomic.AddInt32(&p.txCount, 1)
	if len(p.txSet) == p.commandSize {
		p.seqNo++
		command := types.GenerateCommand(p.author, p.seqNo, p.txSet)

		//if command.Sequence%uint64(p.interval) == 0 {
		//	p.timer.startTimerOnlyOne(command)
		//	command = p.frontInfo(command)
		//}

		p.sender.BroadcastCommand(command)
		p.logger.Infof("[%d] generate command %s", p.author, command.Format())
		p.txSet = nil
	}
}

func (p *proposerImpl) frontInfo(command *protos.Command) *protos.Command {
	command.FrontRunner = p.front.read(p.readFront)
	return command
}
