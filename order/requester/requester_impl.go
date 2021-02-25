package requester

import (
	teeTypes "github.com/Grivn/phalanx-common/tee/types"
	commonTypes "github.com/Grivn/phalanx-common/types"
	commonProto "github.com/Grivn/phalanx-common/types/protos"
	"github.com/Grivn/phalanx-order/external"
	"github.com/Grivn/phalanx-order/requester/types"

	logger "github.com/ultramesh/fancylogger"

	"github.com/golang/protobuf/proto"
)

type requesterImpl struct {
	author uint64

	batchSize int

	// a buffer channel
	txc chan *commonProto.Transaction

	reqC chan *commonProto.OrderedReq

	close chan bool

	txList []*commonProto.Transaction

	tee external.USIG

	sender external.Network

	sequencePool external.SequencePool

	logger *logger.Logger
}

func newRequesterImpl(config types.Config) *requesterImpl {
	return &requesterImpl{
		author:       config.Author,
		batchSize:    config.BatchSize,
		txc:          config.TxC,
		reqC:         config.ReqC,
		close:        make(chan bool),
		txList:       nil,
		sender:       config.Sender,
		tee:          config.TEE,
		sequencePool: config.SequencePool,
		logger:       config.Logger,
	}
}

func (ri *requesterImpl) start() {
	go ri.listener()
}

func (ri *requesterImpl) stop() {
	close(ri.close)
}

func (ri *requesterImpl) listener() {
	for {
		select {
		case tx := <-ri.txc:
			ri.processTransaction(tx)
		case <-ri.close:
			ri.logger.Notice("exist requester listener")
			return
		}
	}
}

func (ri *requesterImpl) processTransaction(tx *commonProto.Transaction) {
	ri.txList = append(ri.txList, tx)

	if len(ri.txList) == ri.batchSize {
		hash := commonTypes.GenerateTxListHash(ri.txList)
		usig, err := ri.tee.CreateUI(hash)

		if err != nil {
			ri.logger.Errorf("Create unique identifier failed: %s", err)
			return
		}

		msg := &commonProto.OrderedMsg{
			Type: commonProto.Type_ORDERED_REQ,
			Author: ri.author,
			BatchHash: commonTypes.BytesToString(hash),
			UI: usig,
		}

		ri.txList = nil

		payload, err := proto.Marshal(request)
		if err != nil {
			ri.logger.Error("Marshal Error: %s", err)
			return
		}
		ri.sender.Broadcast(payload)

		ri.sequencePool.ReqRecorder(request)

		ri.reqC <- request
	}
}
