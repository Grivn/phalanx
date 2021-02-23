package phalanx

import commonProto "github.com/Grivn/phalanx-common/types/protos"

type channelGroup struct {
	txsChan chan *commonProto.Transaction
	reqChan chan *commonProto.OrderedReq
	logChan chan *commonProto.OrderedLog
}

func initChannelGroup(txsC chan *commonProto.Transaction, reqC chan *commonProto.OrderedReq, logC chan *commonProto.OrderedLog) channelGroup {
	return channelGroup{
		txsChan: txsC,
		reqChan: reqC,
		logChan: logC,
	}
}
