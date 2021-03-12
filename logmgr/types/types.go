package types

const (
	LogRecvGenerate = iota
	LogRecvRecord
	LogRecvReady
)

type RecvEvent struct {
	EventType int
	Event     interface{}
}

const (
	LogReplyQuorumBinaryEvent = iota
	LogReplyExecuteEvent
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}
