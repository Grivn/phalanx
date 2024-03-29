package main

import (
	"fmt"
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
	byzRange := 0
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
		logger := types.NewRawLoggerFile(logDir + "/bft-node-" + strconv.Itoa(i+1) + ".log")
		loggerExec := types.NewRawLoggerFile(logDir + "/bft-node-" + strconv.Itoa(i+1) + ".exec.log")
		id := uint64(i + 1)
		exec := mocks.NewSimpleExecutor(id, types.NewRawLogger(), loggerExec)
		byz := false
		if id <= uint64(byzRange) {
			byz = true
		}
		privKey, pubKeys, err := mocks.GenerateKeys(id, n)
		if err != nil {
			panic(fmt.Sprintf("generate keys error: %s", err))
		}
		conf := phalanx.Config{
			Author:      id,
			OLeader:     oLeader,
			Byz:         byz,
			OpenLatency: 0,
			Duration:    types.DefaultTimeDuration,
			Interval:    types.DefaultInterval,
			CDuration:   types.DefaultTimeDuration,
			N:           n,
			Multi:       types.DefaultMulti,
			LogCount:    types.DefaultLogCount,
			MemSize:     types.DefaultMemSize,
			CommandSize: types.SingleCommandSize,
			Selected:    1,
			PrivateKey:  privKey,
			PublicKeys:  pubKeys,
			Exec:        exec,
			Network:     net,
			Logger:      logger,
		}
		phx[id] = phalanx.NewPhalanxProvider(conf)
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
