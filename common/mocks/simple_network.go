package mocks

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/gogo/protobuf/proto"
	"math/rand"
	"time"
)

type SimpleNetwork struct {
	async    bool
	networkC map[uint64]chan *protos.ConsensusMessage
	logger   external.Logger
}

func NewSimpleNetwork(networkC map[uint64]chan *protos.ConsensusMessage, async bool) *SimpleNetwork {
	return &SimpleNetwork{async: async, networkC: networkC, logger: NewRawLogger()}
}

func (net *SimpleNetwork) PhalanxBroadcast(message *protos.ConsensusMessage) {
	if net.async {
		// NOTE: phalanx itself could be running in a asynchronous network environment.
		i := rand.Int()%10
		time.Sleep(time.Duration(i) * time.Millisecond)
	}

	go net.broadcast(message)
}

func (net *SimpleNetwork) PhalanxUnicast(message *protos.ConsensusMessage) {
	go net.unicast(message)
}

func (net *SimpleNetwork) broadcast(message *protos.ConsensusMessage) {
	for _, ch := range net.networkC {
		ch <- message
	}
}

func (net *SimpleNetwork) unicast(message *protos.ConsensusMessage) {
	net.networkC[message.To] <- message
}

func SimpleListener(mgr internal.LogManager, net chan *protos.ConsensusMessage, closeC chan bool) {
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
				qc := &protos.QuorumCert{}
				if err := proto.Unmarshal(msg.Payload, qc); err != nil {
					panic(err)
				}
				if err := mgr.ProcessQC(qc); err != nil {
				panic(err)
			}
			case protos.MessageType_VOTE:
				vote := &protos.Vote{}
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
