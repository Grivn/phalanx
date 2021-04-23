package types

import commonProto "github.com/Grivn/phalanx/common/types/protos"

// CommChan is used to receive the communicate messages send from internet
type CommChan struct {
	BatchChan chan *commonProto.Batch

	ReqChan chan *commonProto.OrderedReq

	LogChan chan *commonProto.OrderedLog

	AckChan chan *commonProto.OrderedAck
}

// TxPoolRecvChan is the channel group which is used to receive events from other modules for txPool
type TxPoolRecvChan struct {
	// TransactionChan is used to receive transactions send from api
	TransactionChan chan *commonProto.Transaction

	// BatchedChan is used to receive batchStore send from other replicas
	BatchedChan chan *commonProto.Batch

	// ExecuteChan is used to receive execute event send from executor module
	ExecuteChan chan *Block
}

// TxPoolSendChan is the channel group which is used to send back information to other modules for txPool
type TxPoolSendChan struct {
	// BatchedChan is used to send back the batch generated by txPool
	BatchedChan chan *commonProto.Batch
}

// RequesterRecvChan is the channel group which is used to receive events from other modules for requester
type RequesterRecvChan struct {
	OrderedChan chan *commonProto.OrderedReq
}

// RequesterSendChan is the channel group which is used to send back information to other modules for requester
type RequesterSendChan struct {
	BatchIdChan chan *commonProto.BatchId
}

// ReliableRecvChan is the channel group which is used to receive events from other modules for reliable-logs
type ReliableRecvChan struct {
	// BatchIdChan is used to receive the batch's identifier provided by requester module
	BatchIdChan chan *commonProto.BatchId

	// LogChan is used to receive the ordered-log generated by self or send from other replicas
	LogChan chan *commonProto.OrderedLog

	// AckChan is used to receive the ordered-ack generated by self or send from other replicas
	AckChan chan *commonProto.OrderedAck
}

// ReliableSendChan is the channel group which is used to send back information to other modules for reliable-logs
type ReliableSendChan struct {
	// TrustedChan is used to send back the trusted ordered-log (f+1 copies)
	TrustedChan chan *commonProto.OrderedLog

	// StableChan is used to send back the stable ordered-log (n-f copies)
	StableChan chan *commonProto.OrderedLog
}

// ExecutorRecvChan is the channel group which is used to receive events from other modules for executor
type ExecutorRecvChan struct {
	// ExecuteLogsChan is used to receive the logs which could be executed
	ExecuteLogsChan chan *commonProto.ExecuteLogs
}

// ExecutorSendChan is the channel group which is used to send back information to other modules for executor
type ExecutorSendChan struct {
	BlockChan chan *Block
}
