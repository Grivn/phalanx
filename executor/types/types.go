package types

import "github.com/Grivn/phalanx/common/types/protos"

const (
	ExecRecvLogs = iota
	ExecRecvBatch
)

type ExecuteLogs struct {
	Sequence uint64
	Logs     []*protos.OrderedMsg
}

type RecvEvent struct {
	EventType int
	Event     interface{}
}

const (
	ExecReplyLoadBatch = iota
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}
