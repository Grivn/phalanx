package logmanager

import (
	"testing"
	"time"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/sequencepool"
)

func TestLogManager(t *testing.T) {
	_ = crypto.SetKeys()

	n := 4

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		nc[id] = make(chan *protos.ConsensusMessage)
	}

	net := mocks.NewSimpleNetwork(nc)

	lms := make(map[uint64]internal.LogManager)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		sp := sequencepool.NewSequencePool(id, n)
		lms[id] = NewLogManager(n, id, sp, net, mocks.NewRawLogger())
	}

	closeC := make(chan bool)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		go mocks.SimpleListener(lms[id], nc[id], closeC)
	}

	command := mocks.NewCommand()
	_ = lms[uint64(1)].ProcessCommand(command)

	time.Sleep(time.Second)
}
