package types

const (
	RecvRecordTxEvent = iota
	RecvRecordBatchEvent
	RecvExecuteBlock
)

type RecvEvent struct {
	EventType int
	Event     interface{}
}

const (
	ReplyGenerateBatchEvent = iota
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}
