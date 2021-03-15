package types

const (
	BinaryRecvTag = iota
	BinaryRecvNotification
)

type RecvEvent struct {
	EventType uint64
	Event     interface{}
}

const (
	BinaryReplyReady = iota
)

type ReplyEvent struct {
	EventType uint64
	Event     interface{}
}
