package test

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/core"
)

func TestPhalanx(t *testing.T) {
	_ = crypto.SetKeys()

	n := 4

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	closeC := make(chan bool)

	phx := make(map[uint64]phalanx.SynchronousProvider)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		nc[id] = make(chan *protos.ConsensusMessage)
	}
	net := mocks.NewSimpleNetwork(nc)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		exec := mocks.NewSimpleExecutor(id, mocks.NewRawLogger())
		phx[id] = phalanx.NewPhalanxProvider(n, id, exec, net, mocks.NewRawLoggerFile("node-"+strconv.Itoa(i+1)))
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		go phalanxListener(phx[id], nc[id], closeC)
	}

	replicas := make(map[uint64]*replica)
	bftCs := make(map[uint64]chan *bftMessage)
	sendC := make(chan *bftMessage)
	for i:=0; i<4; i++ {
		id := uint64(i+1)
		bftCs[id] = make(chan *bftMessage)
		replicas[id] = newReplica(n, id, phx[id], sendC, bftCs[id], closeC, mocks.NewRawLoggerFile("bft-node-"+strconv.Itoa(i+1)))
		replicas[id].run()
	}
	go cluster(sendC, bftCs, closeC)

	count := 2000
	for i:=0; i<count; i++ {
		go commandSender(phx)
	}

	time.Sleep(3000 * time.Second)
}

func phalanxListener(phx phalanx.SynchronousProvider, net chan *protos.ConsensusMessage, closeC chan bool) {
	for {
		select {
		case message := <-net:
			phx.ProcessConsensusMessage(message)
		case <-closeC:
			return
		}
	}
}

func commandSender(phx map[uint64]phalanx.SynchronousProvider) {
	i := rand.Int()%10
	time.Sleep(time.Duration(i) * time.Millisecond)

	command := mocks.NewCommand()

	for _, p := range phx {
		go p.ProcessCommand(command)
	}
}
