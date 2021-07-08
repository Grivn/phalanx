package test

import (
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

	cc := make(map[uint64]chan *protos.Command)

	closeC := make(chan bool)

	phx := make(map[uint64]phalanx.Provider)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		nc[id] = make(chan *protos.ConsensusMessage)
		cc[id] = make(chan *protos.Command)
	}
	net := mocks.NewSimpleNetwork(nc, cc, types.NewRawLogger(), async)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		exec := mocks.NewSimpleExecutor(id, types.NewRawLogger())
		phx[id] = phalanx.NewPhalanxProvider(n, id, types.DefaultLogRotation, types.DefaultTimeDuration,
			exec, net, types.NewRawLogger())
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		go phalanxListener(phx[id], nc[id], cc[id], closeC)
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

	num := 2000
	client := 4
	transactionSendInstance(num, client, phx)
	//commandSendInstance(num, client, phx)

	time.Sleep(3000 * time.Second)
}

func transactionSendInstance(num, client int, phx map[uint64]phalanx.Provider) {
	for c:=0; c<client; c++ {
		for i:=0; i<num; i++ {
			go transactionSender(uint64(c+1), phx)
		}
	}
}

func commandSendInstance(num, client int, phx map[uint64]phalanx.Provider) {
	for c:=0; c<client; c++ {
		for i:=0; i<num; i++ {
			go commandSender(uint64(c+1), uint64(i+1), phx)
		}
	}
}

func phalanxListener(phx phalanx.Provider, net chan *protos.ConsensusMessage, cmd chan *protos.Command, closeC chan bool) {
	for {
		select {
		case message := <-net:
			phx.ProcessConsensusMessage(message)
		case command := <-cmd:
			go phx.ProcessCommand(command)
		case <-closeC:
			return
		}
	}
}

func transactionSender(sender uint64, phx map[uint64]phalanx.Provider) {
	tx := types.GenerateRandTransaction(1)

	go phx[sender].ProcessTransaction(tx)
}

func commandSender(sender, seqNo uint64, phx map[uint64]phalanx.Provider) {
	command := types.GenerateRandCommand(sender, seqNo, 1, 1)

	for _, p := range phx {
		go p.ProcessCommand(command)
	}
}
