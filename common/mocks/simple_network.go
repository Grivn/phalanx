package mocks

import (
	"math/rand"
	"time"

	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"

	"github.com/gogo/protobuf/proto"
)

type SimpleNetwork struct {
	async    bool
	networkC map[uint64]chan *protos.PConsensusMessage
	commandC map[uint64]chan *protos.PCommand
	logger   external.Logger
}

func NewSimpleNetwork(networkC map[uint64]chan *protos.PConsensusMessage, commandC map[uint64]chan *protos.PCommand, logger external.Logger, async bool) *SimpleNetwork {
	return &SimpleNetwork{async: async, networkC: networkC, commandC: commandC, logger: logger}
}

func (net *SimpleNetwork) BroadcastPCM(message *protos.PConsensusMessage) {
	if net.async {
		// NOTE: phalanx itself could be running in a asynchronous network environment.
		i := rand.Int()%10
		time.Sleep(time.Duration(i) * time.Millisecond)
	}

	go net.broadcast(message)
}

func (net *SimpleNetwork) UnicastPCM(message *protos.PConsensusMessage) {
	go net.unicast(message)
}

func (net *SimpleNetwork) BroadcastCommand(command *protos.PCommand) {
	if net.async {
		// NOTE: phalanx itself could be running in a asynchronous network environment.
		i := rand.Int()%10
		time.Sleep(time.Duration(i) * time.Millisecond)
	}

	go net.sendCommand(command)
}

func (net *SimpleNetwork) broadcast(message *protos.PConsensusMessage) {
	for _, ch := range net.networkC {
		ch <- message
	}
}

func (net *SimpleNetwork) sendCommand(command *protos.PCommand) {
	for _, ch := range net.commandC {
		ch <- command
	}
}

func (net *SimpleNetwork) unicast(message *protos.PConsensusMessage) {
	net.networkC[message.To] <- message
}

func SimpleListener(mgr internal.LogManager, net chan *protos.PConsensusMessage, closeC chan bool) {
	for {
		select {
		case msg := <-net:
			switch msg.Type {
			case protos.MessageType_PRE_ORDER:
				pre := &protos.PreOrder{}
				if err := proto.Unmarshal(msg.Payload, pre); err != nil {
					panic(err)
				}
				if err := mgr.ProcessPreOrder(pre); err != nil {
					panic(err)
				}
			case protos.MessageType_QUORUM_CERT:
				pOrder := &protos.PartialOrder{}
				if err := proto.Unmarshal(msg.Payload, pOrder); err != nil {
					panic(err)
				}
				if err := mgr.ProcessPartial(pOrder); err != nil {
				panic(err)
			}
			case protos.MessageType_VOTE:
				vote := &protos.PVote{}
				if err := proto.Unmarshal(msg.Payload, vote); err != nil {
					panic(err)
				}
				if err := mgr.ProcessVote(vote); err != nil {
				panic(err)
			}
			}
		case <-closeC:
			return
		}
	}
}
