package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Grivn/phalanx"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/mocks"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
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
	sender := mocks.NewSimpleNetwork(nc, cc, types.NewRawLogger(), async)
	for i := 0; i < n; i++ {
		logger := types.NewRawLoggerFile(logDir + "/bft-node-" + strconv.Itoa(i+1) + ".log")
		loggerExec := types.NewRawLoggerFile(logDir + "/bft-node-" + strconv.Itoa(i+1) + ".exec.log")
		id := uint64(i + 1)
		executor := mocks.NewSimpleExecutor(id, types.NewRawLogger(), loggerExec)
		byz := false
		if id <= uint64(byzRange) {
			byz = true
		}
		privateKey, publicKeys, err := mocks.GenerateExternalCrypto(id, n)
		if err != nil {
			panic(fmt.Sprintf("generate keys error: %s", err))
		}
		conf := config.PhalanxConf{
			OligarchID:  oLeader,
			IsByzantine: byz,
			NodeID:      id,
			NodeCount:   n,
			Timeout:     types.DefaultTimeDuration,
			Multi:       types.DefaultMulti,
			MemSize:     types.DefaultMemSize,
			CommandSize: types.SingleCommandSize,
			Selected:    1,
		}
		phx[id] = phalanx.NewPhalanxProvider(conf, privateKey, publicKeys, executor, sender, logger)
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
