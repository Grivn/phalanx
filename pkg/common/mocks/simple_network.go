package mocks

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/external"
	"math/rand"
	"time"
)

type SimpleNetwork struct {
	async    bool
	networkC map[uint64]chan *protos.ConsensusMessage
	commandC map[uint64]chan *protos.Command
	attemptC map[uint64]chan *protos.OrderAttempt
	logger   external.Logger
}

func NewSimpleNetwork(networkC map[uint64]chan *protos.ConsensusMessage, commandC map[uint64]chan *protos.Command, logger external.Logger, async bool) *SimpleNetwork {
	return &SimpleNetwork{async: async, networkC: networkC, commandC: commandC, logger: logger}
}

func (net *SimpleNetwork) BroadcastPCM(message *protos.ConsensusMessage) {
	if net.async {
		// NOTE: phalanx itself could be running in a asynchronous network environment.
		i := rand.Int() % 10
		time.Sleep(time.Duration(i) * time.Millisecond)
	}

	go net.broadcast(message)
}

func (net *SimpleNetwork) UnicastPCM(message *protos.ConsensusMessage) {
	go net.unicast(message)
}

func (net *SimpleNetwork) BroadcastCommand(command *protos.Command) {
	if net.async {
		// NOTE: phalanx itself could be running in a asynchronous network environment.
		i := rand.Int() % 10
		time.Sleep(time.Duration(i) * time.Millisecond)
	}

	go net.sendCommand(command)
}

func (net *SimpleNetwork) broadcast(message *protos.ConsensusMessage) {
	for _, ch := range net.networkC {
		ch <- message
	}
}

func (net *SimpleNetwork) sendCommand(command *protos.Command) {
	for _, ch := range net.commandC {
		ch <- command
	}
}

func (net *SimpleNetwork) sendOrderAttempt(attempt *protos.OrderAttempt) {
	for _, ch := range net.attemptC {
		ch <- attempt
	}
}

func (net *SimpleNetwork) unicast(message *protos.ConsensusMessage) {
	net.networkC[message.To] <- message
}
