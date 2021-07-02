package test

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/core"
)

func TestPhalanx(t *testing.T) {

	n := 4

	async := false

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	closeC := make(chan bool)

	phx := make(map[uint64]phalanx.Provider)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		nc[id] = make(chan *protos.ConsensusMessage)
	}
	net := mocks.NewSimpleNetwork(nc, types.NewRawLogger(), async)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		exec := mocks.NewSimpleExecutor(id, types.NewRawLogger())
		phx[id] = phalanx.NewPhalanxProvider(n, id, types.DefaultLogRotation, types.DefaultTimeDuration,
			exec, net, types.NewRawLogger())
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		go phalanxListener(phx[id], nc[id], closeC)
	}

	replicas := make(map[uint64]*replica)
	bftCs := make(map[uint64]chan *bftMessage)
	sendC := make(chan *bftMessage)
	logDir := "bft_nodes_"+time.Now().Format("2006-01-02_15:04:05")
	_ = os.Mkdir(logDir, os.ModePerm)
	for i:=0; i<4; i++ {
		id := uint64(i+1)
		bftCs[id] = make(chan *bftMessage)
		replicas[id] = newReplica(n, id, phx[id], sendC, bftCs[id], closeC, types.NewRawLoggerFile(logDir+"/bft-node-"+strconv.Itoa(i+1)+".log"))
		replicas[id].run()
	}
	go cluster(sendC, bftCs, closeC)

	count := 1000
	for i:=0; i<count; i++ {
		go commandSender(phx)
	}

	time.Sleep(3000 * time.Second)
}

func phalanxListener(phx phalanx.Provider, net chan *protos.ConsensusMessage, closeC chan bool) {
	for {
		select {
		case message := <-net:
			phx.ProcessConsensusMessage(message)
		case <-closeC:
			return
		}
	}
}

func commandSender(phx map[uint64]phalanx.Provider) {
	i := rand.Int()%10
	time.Sleep(time.Duration(i) * time.Millisecond)

	command := types.GenerateRandCommand(200, 5)

	for _, p := range phx {
		go p.ProcessCommand(command)
	}
}
