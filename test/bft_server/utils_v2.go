package main

import (
	"time"

	"github.com/Grivn/phalanx"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
)

//nolint
func transactionSendInstanceV2(num, client int, phx map[uint64]phalanx.ProviderV2) {
	for i := 0; i < num; i++ {
		time.Sleep(2 * time.Microsecond)
		for c := 0; c < client; c++ {
			go transactionSenderV2(uint64(c+1), phx)
		}
	}
}

func commandSendInstanceV2(num, client int, phx map[uint64]phalanx.ProviderV2) {
	for i := 0; i < num; i++ {
		time.Sleep(2 * time.Microsecond)
		for c := 0; c < client; c++ {
			go commandSenderV2(uint64(c+1), uint64(i+1), phx)
		}
	}
}

func phalanxListenerV2(phx phalanx.ProviderV2, net chan *protos.ConsensusMessage, cmd chan *protos.Command, closeC chan bool) {
	for {
		select {
		case message := <-net:
			phx.ReceiveConsensusMessage(message)
		case command := <-cmd:
			phx.ReceiveCommand(command)
		case <-closeC:
			return
		}
	}
}

//nolint
func transactionSenderV2(sender uint64, phx map[uint64]phalanx.ProviderV2) {
	tx := types.GenerateRandTransaction(1)

	phx[sender].ReceiveTransaction(tx)
}

func commandSenderV2(sender, seqNo uint64, phx map[uint64]phalanx.ProviderV2) {
	command := types.GenerateRandCommand(sender, seqNo, 1, 1)

	for _, p := range phx {
		go p.ReceiveCommand(command)
	}
}
