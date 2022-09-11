package receiver

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"time"
)

type buyerImpl struct {
	id uint64

	selected uint64

	itemNo uint64

	timer *localTimer

	snappingUpC chan bool

	closeC chan bool

	sender external.NetworkService

	logger external.Logger
}

func newBuyer(id uint64, conf Config) *buyerImpl {
	snappingUpC := make(chan bool)
	return &buyerImpl{
		id:          id,
		itemNo:      uint64(0),
		timer:       newLocalTimer(snappingUpC),
		snappingUpC: snappingUpC,
		closeC:      make(chan bool),
		selected:    conf.Selected,
		sender:      conf.Sender,
		logger:      conf.Logger,
	}
}

func (b *buyerImpl) run() {
	if b.id == uint64(2) {
		time.Sleep(500 * time.Millisecond)
	}
	go b.listener()
	b.timer.updateDuration(b.generateDuration())
	b.timer.startTimer()
}

func (b *buyerImpl) quit() {
	select {
	case <-b.closeC:
	default:
		close(b.closeC)
	}
}

func (b *buyerImpl) listener() {
	for {
		select {
		case <-b.snappingUpC:
			b.snappingUp()
			b.timer.updateDuration(b.generateDuration())
			b.timer.startTimer()
		default:
			continue
		}
	}
}

func (b *buyerImpl) snappingUp() {
	if b.selected != uint64(0) && b.id > b.selected {
		return
	}
	b.itemNo++
	command := types.GenerateCommand(b.id, b.itemNo, nil)
	b.logger.Infof("[%d] generate command %s", b.id, command.FormatSnappingUp())
	b.sender.BroadcastCommand(command)
}

func (b *buyerImpl) generateDuration() time.Duration {
	now := time.Now()
	next := now.Add(time.Second * 5)
	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location())
	return next.Sub(now)
}
