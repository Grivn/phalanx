package mocks

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type network struct {
	author uint64
	commC  map[uint64]commonTypes.CommChan
	logger external.Logger
}

func NewFakeNetwork(author uint64, commC map[uint64]commonTypes.CommChan) external.Network {
	logger := NewRawLogger()

	net := &network{
		author: author,
		commC:  commC,
		logger: logger,
	}

	return net
}

func (network *network) BroadcastBatch(batch *commonProto.Batch) {
	for id := range network.commC {
		if id == network.author {
			continue
		}
		go func(id uint64) {
			network.commC[id].BatchChan <- batch
		}(id)
	}
}

func (network *network) BroadcastReq(req *commonProto.OrderedReq) {
	for id := range network.commC {
		if id == network.author {
			continue
		}
		go func(id uint64) {
			network.commC[id].ReqChan <- req
		}(id)
	}
}

func (network *network) BroadcastLog(log *commonProto.OrderedLog) {
	for id := range network.commC {
		if id == network.author {
			continue
		}
		go func(id uint64) {
			network.commC[id].LogChan <- log
		}(id)
	}
}

func (network *network) BroadcastAck(ack *commonProto.OrderedAck) {
	for id := range network.commC {
		if id == network.author {
			continue
		}
		go func(id uint64) {
			network.commC[id].AckChan <- ack
		}(id)
	}
}
