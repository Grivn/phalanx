package types

import "github.com/Grivn/phalanx/common/types/protos"

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
	LogReplyMissingEvent
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}

type MissingMsg struct {
	Tag       *protos.BinaryTag
	MissingID []uint64
}
