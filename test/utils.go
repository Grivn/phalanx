package main

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	phalanx "github.com/Grivn/phalanx/core"
	"time"
)

func transactionSendInstance(num, client int, phx map[uint64]phalanx.Provider) {
	for i:=0; i<num; i++ {
		time.Sleep(2*time.Microsecond)
		for c:=0; c<client; c++ {
			go transactionSender(uint64(c+1), phx)
		}
	}
}

func commandSendInstance(num, client int, phx map[uint64]phalanx.Provider) {
	for i:=0; i<num; i++ {
		time.Sleep(2*time.Microsecond)
		for c:=0; c<client; c++ {
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
			phx.ProcessCommand(command)
		case <-closeC:
			return
		}
	}
}

func transactionSender(sender uint64, phx map[uint64]phalanx.Provider) {
	tx := types.GenerateRandTransaction(1)

	phx[sender].ProcessTransaction(tx)
}

func commandSender(sender, seqNo uint64, phx map[uint64]phalanx.Provider) {
	command := types.GenerateRandCommand(sender, seqNo, 1, 1)

	for _, p := range phx {
		go p.ProcessCommand(command)
	}
}
