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

func phalanxRunnerV2() {
	n := 4
	byzRange := 0
	oLeader := uint64(0)

	async := false

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	cc := make(map[uint64]chan *protos.Command)

	closeC := make(chan bool)

	phx := make(map[uint64]phalanx.ProviderV2)

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
		phx[id] = phalanx.NewPhalanxV2(conf, privateKey, publicKeys, executor, sender, logger)
		phx[id].Run()
	}

	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		go phalanxListenerV2(phx[id], nc[id], cc[id], closeC)
	}

	replicas := make(map[uint64]*replicaV2)
	bftCs := make(map[uint64]chan *bftMessageV2)
	sendC := make(chan *bftMessageV2)

	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		bftCs[id] = make(chan *bftMessageV2)
		replicas[id] = newReplicaV2(n, id, phx[id], sendC, bftCs[id], closeC, types.NewRawLoggerFile(logDir+"/bft-node-"+strconv.Itoa(i+1)+".log"))
		replicas[id].run()
	}
	go clusterV2(sendC, bftCs, closeC)

	num := 1000
	//client := 16
	transactionSendInstanceV2(num, n, phx)
	//commandSendInstanceV2(num, client, phx)

	time.Sleep(1000 * time.Second)
}
