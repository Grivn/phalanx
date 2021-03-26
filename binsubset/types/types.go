package types

const (
	BinaryRecvTag = iota
	BinaryRecvNotification
)

type RecvEvent struct {
	EventType int
	Event     interface{}
}

const (
	BinaryReplyReady = iota
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}
