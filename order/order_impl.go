package order

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/order/external"
	"github.com/Grivn/phalanx/order/internal"
	"github.com/Grivn/phalanx/order/requester"
	"github.com/Grivn/phalanx/order/types"
	"github.com/Grivn/phalanx/sgx"

	requesterTypes "github.com/Grivn/phalanx-order/requester/types"

	logger "github.com/ultramesh/fancylogger"
)

type orderImpl struct {
	author uint64

	requester internal.Requester
	collector internal.Collector

	usig external.USIG

	txChan  chan *commonProto.Transaction
	reqChan chan *commonProto.OrderedReq
	logChan chan *commonProto.OrderedLog
	close   chan bool

	logger *logger.Logger
}

func newOrderImpl(config types.Config) *orderImpl {
	usig, err := sgx.NewUSIG(config.EnclaveFile, config.SealedKey)
	if err != nil {
		config.Logger.Errorf("Failed to create usig: %s", err)
		return nil
	}


	txChan := make(chan *commonProto.Transaction, 1000)
	reqChan := make(chan *commonProto.OrderedReq, 1000)
	logChan := make(chan *commonProto.OrderedLog, 1000)

	rc := requesterTypes.Config{
		Author:       config.Author,
		BatchSize:    config.BatchSize,
		TxC:          txChan,
		ReqC:         reqChan,
		Sender:       config.Network,
		TEE:          config.TEE,
		SequencePool: config.SequencePool,
	}
	r := requester.NewRequester(rc)

	return &orderImpl{
		author: config.Author,
		requester: r,

		usig: usig,

		txChan: txChan,
		reqChan: reqChan,
		logChan: logChan,

		close: make(chan bool),

		logger: config.Logger,
	}
}

func (oi *orderImpl) start() {
	oi.collector.Start()
	oi.requester.Start()
}

func (oi *orderImpl) stop() {
	oi.requester.Stop()
	oi.collector.Stop()
}

func (oi *orderImpl) receiveTransaction(tx *commonProto.Transaction) {
	oi.txChan <- tx
}

func (oi *orderImpl) receiveOrderedReq(req *commonProto.OrderedReq) {
	if req.Author == oi.author {
		oi.logger.Warningf("Replica %d received an ordered req from author equal to self", oi.author)
		return
	}
	oi.reqChan <- req
}

func (oi *orderImpl) receiveOrderedLog(log *commonProto.OrderedLog) {
	if log.Author == oi.author {
		oi.logger.Warningf("Replica %d received an ordered log from author equal to self", oi.author)
		return
	}
	oi.logChan <- log
}
