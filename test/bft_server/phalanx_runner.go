package main

import (
	"os"
	"strconv"
	"time"

	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	phalanx "github.com/Grivn/phalanx/core"
)

func phalanxRunner() {
	n := 4
	byzRange := 1
	oLeader := uint64(0)

	async := false

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	cc := make(map[uint64]chan *protos.Command)

	closeC := make(chan bool)

	phx := make(map[uint64]phalanx.Provider)

	logDir := "bft_nodes"
	_ = os.Mkdir(logDir, os.ModePerm)

	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		nc[id] = make(chan *protos.ConsensusMessage)
		cc[id] = make(chan *protos.Command)
	}
	net := mocks.NewSimpleNetwork(nc, cc, types.NewRawLogger(), async)
	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		exec := mocks.NewSimpleExecutor(id, types.NewRawLogger())
		byz := false
		if id <= uint64(byzRange) {
			byz = true
		}
		phx[id] = phalanx.NewPhalanxProvider(oLeader, byz, 0, types.DefaultTimeDuration, types.DefaultInterval, types.DefaultTimeDuration, n, types.DefaultMulti, types.DefaultLogCount, types.DefaultMemSize, id, types.SingleCommandSize, exec, net, types.NewRawLoggerFile(logDir+"/bft-node-"+strconv.Itoa(i+1)+".log"), 1)
		phx[id].Run()
	}

	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		go phalanxListener(phx[id], nc[id], cc[id], closeC)
	}

	replicas := make(map[uint64]*replica)
	bftCs := make(map[uint64]chan *bftMessage)
	sendC := make(chan *bftMessage)

	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		bftCs[id] = make(chan *bftMessage)
		replicas[id] = newReplica(n, id, phx[id], sendC, bftCs[id], closeC, types.NewRawLoggerFile(logDir+"/bft-node-"+strconv.Itoa(i+1)+".log"))
		replicas[id].run()
	}
	go cluster(sendC, bftCs, closeC)

	num := 1000
	//client := 16
	transactionSendInstance(num, n, phx)
	//commandSendInstance(num, client, phx)

	time.Sleep(1000 * time.Second)
}
