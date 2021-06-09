package phalanx

import (
	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"
	"strconv"
	"time"

	"testing"
)

func init() {
	_ = crypto.SetKeys()
}

func TestPhalanx(t *testing.T) {
	n := 4

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	closeC := make(chan bool)

	phx := make(map[uint64]Provider)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		nc[id] = make(chan *protos.ConsensusMessage)
	}
	net := mocks.NewSimpleNetwork(nc)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		exec := mocks.NewSimpleExecutor(id, mocks.NewRawLogger())
		phx[id] = NewPhalanxProvider(n, id, exec, net, mocks.NewRawLoggerFile("node"+strconv.Itoa(i+1)))
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		go phalanxListener(phx[id], nc[id], closeC)
	}

	commandSender(phx)

	time.Sleep(3 * time.Second)
}

func phalanxListener(phx Provider, net chan *protos.ConsensusMessage, closeC chan bool) {
	for {
		select {
		case message := <-net:
			phx.ProcessConsensusMessage(message)
		case <-closeC:
			return
		}
	}
}

func commandSender(phx map[uint64]Provider) {
	command := mocks.NewCommand()

	for _, p := range phx {
		p.ProcessCommand(command)
	}
}

func bftConsensus(phx map[uint64]Provider) {
	// fixed-leader byzantine fault tolerance consensus simulator

	leader := uint64(1)

	bftCs := make(map[uint64]chan bftMessage)
}

const (
	proposal = iota
	vote
	quorum
)

type replica struct {
	phalanx Provider

	sequence uint64
	aggMap map[uint64]uint64
	bftC chan bftMessage
}

type bftMessage struct {
	from uint64
	to uint64
	sequence uint64
	typ      int
	payload  []byte
}

func (replica *replica) bftListener() {
	for {
		select {
		case msg := <-replica.bftC:
			switch msg.typ {
			case proposal:

			case vote:

			case quorum:

			}
		}
	}
}

func (replica *replica) proposer(id uint64, ph Provider) {
	sequence := uint64(0)
	aggMap := make(map[uint64]uint64)

	// generation of proposals
	go func(id uint64, bftCs map[uint64]chan bftMessage) {
		for {
			payload, err := ph.MakePayload()
			if err != nil {
				continue
			}

			sequence++
			aggMap[sequence]++
			bftC <- bftMessage{typ: proposal, sequence: sequence, payload: payload}
		}
	}(id, bftCs)
}

func leader() {

}

func follower(id uint64, ph Provider, bftCs map[uint64]chan bftMessage) {
	sequence := uint64(0)
	cache := make(map[uint64]bftMessage)

	for {

	}
}
